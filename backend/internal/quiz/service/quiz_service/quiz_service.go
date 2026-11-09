package quiz_service

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"main/internal/pkg/connect_manager"
	"main/internal/pkg/errs"
	"main/internal/pkg/logger"
	"main/internal/quiz/dto/quiz_dto"
	"main/internal/quiz/model/quiz_answer"
	"main/internal/quiz/model/quiz_buzz"
	"main/internal/quiz/model/quiz_question"
	"main/internal/quiz/model/quiz_session"
	"main/internal/quiz/model/quiz_team"
	"main/internal/team/model/team"
	"math/big"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	Round1 = 1
	Round2 = 2
	Round3 = 3

	MaxAttemptsRound1 = 3
	MaxQueueRound1    = 3
)

var RoundConfigs = []quiz_dto.RoundConfig{
	{Round: 1, Name: "Стрелы логики", QuestionCount: 10, TimeSec: 60},
	{Round: 2, Name: "Ударный вектор", QuestionCount: 7, TimeSec: 60},
	{Round: 3, Name: "Формула клинка", QuestionCount: 5, TimeSec: 120},
}

func roundConfig(round int) (quiz_dto.RoundConfig, bool) {
	for _, cfg := range RoundConfigs {
		if cfg.Round == round {
			return cfg, true
		}
	}
	return quiz_dto.RoundConfig{}, false
}

// timerRuntime holds the server-authoritative countdown state for one event.
type timerRuntime struct {
	duration  time.Duration
	remaining time.Duration
	deadline  time.Time
	running   bool
	started   bool
}

func (t *timerRuntime) remainingNow() time.Duration {
	if !t.started {
		return t.duration
	}
	if t.running {
		rem := time.Until(t.deadline)
		if rem < 0 {
			return 0
		}
		return rem
	}
	return t.remaining
}

type QuizSessionRepositoryInterface interface {
	GetOrCreate(ctx context.Context, db *gorm.DB, eventID uuid.UUID) (quiz_session.QuizSession, error)
	Save(ctx context.Context, db *gorm.DB, session quiz_session.QuizSession) error
}

type QuizQuestionRepositoryInterface interface {
	GetByEventAndRound(ctx context.Context, db *gorm.DB, eventID uuid.UUID, round int) ([]quiz_question.QuizQuestion, error)
	GetByEvent(ctx context.Context, db *gorm.DB, eventID uuid.UUID) ([]quiz_question.QuizQuestion, error)
	GetByNumber(ctx context.Context, db *gorm.DB, eventID uuid.UUID, round int, number int) (quiz_question.QuizQuestion, error)
	Upsert(ctx context.Context, db *gorm.DB, question quiz_question.QuizQuestion) error
	DeleteByID(ctx context.Context, db *gorm.DB, id uuid.UUID) error
}

type QuizTeamRepositoryInterface interface {
	GetByEventID(ctx context.Context, db *gorm.DB, eventID uuid.UUID) ([]quiz_team.QuizTeam, error)
	GetByTeamID(ctx context.Context, db *gorm.DB, eventID uuid.UUID, teamID uuid.UUID) (quiz_team.QuizTeam, error)
	GetByAccessCode(ctx context.Context, db *gorm.DB, eventID uuid.UUID, code string) (quiz_team.QuizTeam, error)
	Save(ctx context.Context, db *gorm.DB, team quiz_team.QuizTeam) error
}

type QuizBuzzRepositoryInterface interface {
	GetByQuestion(ctx context.Context, db *gorm.DB, eventID uuid.UUID, round int, questionNumber int) ([]quiz_buzz.QuizBuzz, error)
	GetByTeamAndQuestion(ctx context.Context, db *gorm.DB, eventID uuid.UUID, round int, questionNumber int, teamID uuid.UUID) (quiz_buzz.QuizBuzz, error)
	Save(ctx context.Context, db *gorm.DB, buzz quiz_buzz.QuizBuzz) error
	DeleteByQuestion(ctx context.Context, db *gorm.DB, eventID uuid.UUID, round int, questionNumber int) error
}

type QuizAnswerRepositoryInterface interface {
	GetByTeamAndRound(ctx context.Context, db *gorm.DB, eventID uuid.UUID, round int, teamID uuid.UUID) ([]quiz_answer.QuizAnswer, error)
	Upsert(ctx context.Context, db *gorm.DB, answer quiz_answer.QuizAnswer) error
}

type TeamRepositoryInterface interface {
	GetAcceptTeamsByEventId(ctx context.Context, db *gorm.DB, eventID uuid.UUID) ([]team.Team, error)
}

type QuizService struct {
	log *slog.Logger
	db  *gorm.DB

	sessionRepository  QuizSessionRepositoryInterface
	questionRepository QuizQuestionRepositoryInterface
	teamRepository     QuizTeamRepositoryInterface
	buzzRepository     QuizBuzzRepositoryInterface
	answerRepository   QuizAnswerRepositoryInterface
	eventTeamRepository TeamRepositoryInterface

	hub *connect_manager.Hub

	mu     sync.Mutex
	timers map[uuid.UUID]*timerRuntime
}

func New(log *slog.Logger, db *gorm.DB,
	sessionRepository QuizSessionRepositoryInterface,
	questionRepository QuizQuestionRepositoryInterface,
	teamRepository QuizTeamRepositoryInterface,
	buzzRepository QuizBuzzRepositoryInterface,
	answerRepository QuizAnswerRepositoryInterface,
	eventTeamRepository TeamRepositoryInterface,
	hub *connect_manager.Hub) *QuizService {
	return &QuizService{
		log:                 log,
		db:                  db,
		sessionRepository:   sessionRepository,
		questionRepository:  questionRepository,
		teamRepository:      teamRepository,
		buzzRepository:      buzzRepository,
		answerRepository:    answerRepository,
		eventTeamRepository: eventTeamRepository,
		hub:                 hub,
		timers:              make(map[uuid.UUID]*timerRuntime),
	}
}

// ---------- helpers ----------

func generateAccessCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return fmt.Sprintf("%04d", time.Now().UnixNano()%10000)
	}
	return fmt.Sprintf("%04d", n.Int64())
}

func (s *QuizService) getSession(ctx context.Context, eventID uuid.UUID) (quiz_session.QuizSession, error) {
	return s.sessionRepository.GetOrCreate(ctx, s.db, eventID)
}

func (s *QuizService) getTimer(eventID uuid.UUID, cfg quiz_dto.RoundConfig) *timerRuntime {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.timers[eventID]
	if !ok {
		t = &timerRuntime{duration: time.Duration(cfg.TimeSec) * time.Second, remaining: time.Duration(cfg.TimeSec) * time.Second}
		s.timers[eventID] = t
	}
	return t
}

// activeBuzz returns the buzz that currently holds the answering turn:
// the first pending buzz within the queue window.
func activeBuzz(buzzes []quiz_buzz.QuizBuzz) *quiz_buzz.QuizBuzz {
	pending := 0
	for i := range buzzes {
		if buzzes[i].Status != quiz_buzz.StatusPending {
			continue
		}
		if pending < MaxQueueRound1 {
			return &buzzes[i]
		}
		pending++
	}
	return nil
}

func attemptsUsed(buzzes []quiz_buzz.QuizBuzz) int {
	used := 0
	for _, b := range buzzes {
		if b.Status == quiz_buzz.StatusWrong || b.Status == quiz_buzz.StatusCorrect {
			used++
		}
	}
	return used
}

// ---------- setup & questions ----------

func (s *QuizService) SetupTeams(ctx context.Context, eventID uuid.UUID) error {
	const op = "service.QuizService.SetupTeams"
	log := s.log.With(slog.String("op", op))

	teams, err := s.eventTeamRepository.GetAcceptTeamsByEventId(ctx, s.db, eventID)
	if err != nil {
		log.Error("failed get accepted teams", logger.Err(err))
		return errs.ErrInternalError
	}

	existing, err := s.teamRepository.GetByEventID(ctx, s.db, eventID)
	if err != nil {
		log.Error("failed get quiz teams", logger.Err(err))
		return errs.ErrInternalError
	}
	existingByTeam := make(map[uuid.UUID]bool, len(existing))
	for _, qt := range existing {
		existingByTeam[qt.TeamID] = true
	}

	for _, t := range teams {
		if existingByTeam[t.ID] {
			continue
		}
		code := generateAccessCode()
		// extremely unlikely collision guard
		for {
			if _, err := s.teamRepository.GetByAccessCode(ctx, s.db, eventID, code); err != nil {
				break
			}
			code = generateAccessCode()
		}
		qt := quiz_team.QuizTeam{
			EventID:    eventID,
			TeamID:     t.ID,
			AccessCode: code,
		}
		if err := s.teamRepository.Save(ctx, s.db, qt); err != nil {
			log.Error("failed save quiz team", logger.Err(err))
			return errs.ErrInternalError
		}
	}

	return s.broadcastState(ctx, eventID)
}

// ListTeams returns quiz teams including access codes (admin only).
func (s *QuizService) ListTeams(ctx context.Context, eventID uuid.UUID) ([]quiz_dto.AdminTeam, error) {
	const op = "service.QuizService.ListTeams"
	log := s.log.With(slog.String("op", op))

	teams, err := s.teamRepository.GetByEventID(ctx, s.db, eventID)
	if err != nil {
		log.Error("failed get quiz teams", logger.Err(err))
		return nil, errs.ErrInternalError
	}

	result := make([]quiz_dto.AdminTeam, 0, len(teams))
	for _, t := range teams {
		result = append(result, quiz_dto.AdminTeam{
			ID:               t.ID.String(),
			TeamID:           t.TeamID.String(),
			Name:             t.Team.Name,
			AccessCode:       t.AccessCode,
			Round1Score:      t.Round1Score,
			Round2Score:      t.Round2Score,
			Round3Score:      t.Round3Score,
			Round3Order:      t.Round3Order,
			Round3WrongCount: t.Round3WrongCount,
			Round3Eliminated: t.Round3Eliminated,
		})
	}
	return result, nil
}

func (s *QuizService) ListQuestions(ctx context.Context, eventID uuid.UUID) ([]quiz_question.QuizQuestion, error) {
	const op = "service.QuizService.ListQuestions"
	log := s.log.With(slog.String("op", op))

	questions, err := s.questionRepository.GetByEvent(ctx, s.db, eventID)
	if err != nil {
		log.Error("failed get questions", logger.Err(err))
		return nil, errs.ErrInternalError
	}
	return questions, nil
}

func (s *QuizService) UpsertQuestion(ctx context.Context, eventID uuid.UUID, round int, number int, text string, answer string) error {
	const op = "service.QuizService.UpsertQuestion"
	log := s.log.With(slog.String("op", op))

	if _, ok := roundConfig(round); !ok {
		return errs.ErrBadRequest.Wrap("Некорректный раунд")
	}
	if text == "" {
		return errs.ErrBadRequest.Wrap("Текст вопроса не может быть пустым")
	}

	q := quiz_question.QuizQuestion{
		EventID:        eventID,
		Round:          round,
		QuestionNumber: number,
		Text:           text,
		Answer:         answer,
	}
	if err := s.questionRepository.Upsert(ctx, s.db, q); err != nil {
		log.Error("failed upsert question", logger.Err(err))
		return errs.ErrInternalError
	}
	return s.broadcastState(ctx, eventID)
}

func (s *QuizService) DeleteQuestion(ctx context.Context, eventID uuid.UUID, questionID uuid.UUID) error {
	const op = "service.QuizService.DeleteQuestion"
	log := s.log.With(slog.String("op", op))

	if err := s.questionRepository.DeleteByID(ctx, s.db, questionID); err != nil {
		log.Error("failed delete question", logger.Err(err))
		return errs.ErrInternalError
	}
	return s.broadcastState(ctx, eventID)
}

// ---------- flow control (host) ----------

func (s *QuizService) StartRound(ctx context.Context, eventID uuid.UUID, round int) error {
	const op = "service.QuizService.StartRound"
	log := s.log.With(slog.String("op", op))

	cfg, ok := roundConfig(round)
	if !ok {
		return errs.ErrBadRequest.Wrap("Некорректный раунд")
	}

	session, err := s.getSession(ctx, eventID)
	if err != nil {
		log.Error("failed get session", logger.Err(err))
		return errs.ErrInternalError
	}

	session.CurrentRound = round
	session.CurrentQuestion = 0
	session.RoundActive = true
	session.AnswerRevealed = false
	session.LeaderboardShown = false

	if err := s.sessionRepository.Save(ctx, s.db, session); err != nil {
		log.Error("failed save session", logger.Err(err))
		return errs.ErrInternalError
	}

	s.mu.Lock()
	s.timers[eventID] = &timerRuntime{
		duration:  time.Duration(cfg.TimeSec) * time.Second,
		remaining: time.Duration(cfg.TimeSec) * time.Second,
	}
	s.mu.Unlock()

	return s.broadcastState(ctx, eventID)
}

func (s *QuizService) StartQuestion(ctx context.Context, eventID uuid.UUID) error {
	const op = "service.QuizService.StartQuestion"
	log := s.log.With(slog.String("op", op))

	session, err := s.getSession(ctx, eventID)
	if err != nil {
		log.Error("failed get session", logger.Err(err))
		return errs.ErrInternalError
	}
	if !session.RoundActive || session.CurrentRound == 0 {
		return errs.ErrInvalidGameState.Wrap("Раунд не начат")
	}
	cfg, _ := roundConfig(session.CurrentRound)
	if session.CurrentQuestion >= cfg.QuestionCount {
		return errs.ErrInvalidGameState.Wrap("Вопросы этого раунда закончились")
	}

	session.CurrentQuestion++
	session.AnswerRevealed = false
	if err := s.sessionRepository.Save(ctx, s.db, session); err != nil {
		log.Error("failed save session", logger.Err(err))
		return errs.ErrInternalError
	}

	// clear stale buzzes for this question
	if err := s.buzzRepository.DeleteByQuestion(ctx, s.db, eventID, session.CurrentRound, session.CurrentQuestion); err != nil {
		log.Error("failed clear buzzes", logger.Err(err))
		return errs.ErrInternalError
	}

	s.mu.Lock()
	t := s.timers[eventID]
	if t == nil {
		t = &timerRuntime{}
		s.timers[eventID] = t
	}
	t.duration = time.Duration(cfg.TimeSec) * time.Second
	t.remaining = t.duration
	t.deadline = time.Now().Add(t.duration)
	t.running = true
	t.started = true
	s.mu.Unlock()

	return s.broadcastState(ctx, eventID)
}

func (s *QuizService) PauseTimer(ctx context.Context, eventID uuid.UUID) error {
	s.mu.Lock()
	if t, ok := s.timers[eventID]; ok && t.running {
		t.remaining = time.Until(t.deadline)
		if t.remaining < 0 {
			t.remaining = 0
		}
		t.running = false
	}
	s.mu.Unlock()
	return s.broadcastState(ctx, eventID)
}

func (s *QuizService) ResumeTimer(ctx context.Context, eventID uuid.UUID) error {
	s.mu.Lock()
	if t, ok := s.timers[eventID]; ok && t.started && !t.running {
		t.deadline = time.Now().Add(t.remaining)
		t.running = true
	}
	s.mu.Unlock()
	return s.broadcastState(ctx, eventID)
}

func (s *QuizService) RevealAnswer(ctx context.Context, eventID uuid.UUID) error {
	const op = "service.QuizService.RevealAnswer"
	log := s.log.With(slog.String("op", op))

	session, err := s.getSession(ctx, eventID)
	if err != nil {
		log.Error("failed get session", logger.Err(err))
		return errs.ErrInternalError
	}
	session.AnswerRevealed = true
	if err := s.sessionRepository.Save(ctx, s.db, session); err != nil {
		log.Error("failed save session", logger.Err(err))
		return errs.ErrInternalError
	}
	return s.broadcastState(ctx, eventID)
}

// CloseQuestion stops the timer and ends the current question without revealing the answer.
func (s *QuizService) CloseQuestion(ctx context.Context, eventID uuid.UUID) error {
	s.mu.Lock()
	if t, ok := s.timers[eventID]; ok {
		t.remaining = 0
		t.running = false
		t.started = true
	}
	s.mu.Unlock()
	return s.broadcastState(ctx, eventID)
}

func (s *QuizService) ShowLeaderboard(ctx context.Context, eventID uuid.UUID, show bool) error {
	const op = "service.QuizService.ShowLeaderboard"
	log := s.log.With(slog.String("op", op))

	session, err := s.getSession(ctx, eventID)
	if err != nil {
		log.Error("failed get session", logger.Err(err))
		return errs.ErrInternalError
	}
	session.LeaderboardShown = show
	if err := s.sessionRepository.Save(ctx, s.db, session); err != nil {
		log.Error("failed save session", logger.Err(err))
		return errs.ErrInternalError
	}
	return s.broadcastState(ctx, eventID)
}

// GrantTurn manually gives the answering turn to a team (human-error correction).
func (s *QuizService) GrantTurn(ctx context.Context, eventID uuid.UUID, teamID uuid.UUID) error {
	const op = "service.QuizService.GrantTurn"
	log := s.log.With(slog.String("op", op))

	session, err := s.getSession(ctx, eventID)
	if err != nil {
		log.Error("failed get session", logger.Err(err))
		return errs.ErrInternalError
	}
	if session.CurrentRound != Round1 || session.CurrentQuestion == 0 {
		return errs.ErrInvalidGameState.Wrap("Сейчас нет активного вопроса первого раунда")
	}

	buzzes, err := s.buzzRepository.GetByQuestion(ctx, s.db, eventID, session.CurrentRound, session.CurrentQuestion)
	if err != nil {
		log.Error("failed get buzzes", logger.Err(err))
		return errs.ErrInternalError
	}

	var minSeq int64 = 0
	for _, b := range buzzes {
		if minSeq == 0 || b.Seq < minSeq {
			minSeq = b.Seq
		}
	}

	var buzz quiz_buzz.QuizBuzz
	found := false
	for _, b := range buzzes {
		if b.TeamID == teamID {
			buzz = b
			found = true
			break
		}
	}
	if found {
		buzz.Status = quiz_buzz.StatusPending
		buzz.Seq = minSeq - 1
	} else {
		buzz = quiz_buzz.QuizBuzz{
			EventID:        eventID,
			Round:          session.CurrentRound,
			QuestionNumber: session.CurrentQuestion,
			TeamID:         teamID,
			Seq:            minSeq - 1,
			Status:         quiz_buzz.StatusPending,
		}
	}
	if err := s.buzzRepository.Save(ctx, s.db, buzz); err != nil {
		log.Error("failed save buzz", logger.Err(err))
		return errs.ErrInternalError
	}

	s.mu.Lock()
	if t, ok := s.timers[eventID]; ok && t.running {
		t.remaining = time.Until(t.deadline)
		t.running = false
	}
	s.mu.Unlock()

	return s.broadcastState(ctx, eventID)
}

// ResetQuestion clears all buzzes for the current question and resumes the timer.
func (s *QuizService) ResetQuestion(ctx context.Context, eventID uuid.UUID) error {
	const op = "service.QuizService.ResetQuestion"
	log := s.log.With(slog.String("op", op))

	session, err := s.getSession(ctx, eventID)
	if err != nil {
		log.Error("failed get session", logger.Err(err))
		return errs.ErrInternalError
	}
	if session.CurrentQuestion == 0 {
		return errs.ErrInvalidGameState.Wrap("Нет активного вопроса")
	}

	if err := s.buzzRepository.DeleteByQuestion(ctx, s.db, eventID, session.CurrentRound, session.CurrentQuestion); err != nil {
		log.Error("failed clear buzzes", logger.Err(err))
		return errs.ErrInternalError
	}

	s.mu.Lock()
	if t, ok := s.timers[eventID]; ok && t.started && !t.running {
		t.deadline = time.Now().Add(t.remaining)
		t.running = true
	}
	s.mu.Unlock()

	return s.broadcastState(ctx, eventID)
}

// ---------- herald / buzzer ----------

func (s *QuizService) HeraldLogin(ctx context.Context, eventID uuid.UUID, code string) (quiz_dto.HeraldLoginResponse, error) {
	const op = "service.QuizService.HeraldLogin"
	log := s.log.With(slog.String("op", op))

	qt, err := s.teamRepository.GetByAccessCode(ctx, s.db, eventID, code)
	if err != nil {
		log.Error("failed get quiz team by code", logger.Err(err))
		return quiz_dto.HeraldLoginResponse{}, errs.ErrBadRequest.Wrap("Неверный код команды")
	}

	teams, err := s.teamRepository.GetByEventID(ctx, s.db, eventID)
	if err != nil {
		log.Error("failed get quiz teams", logger.Err(err))
		return quiz_dto.HeraldLoginResponse{}, errs.ErrInternalError
	}

	number := 0
	for i, t := range teams {
		if t.TeamID == qt.TeamID {
			number = i + 1
			break
		}
	}

	return quiz_dto.HeraldLoginResponse{
		TeamID: qt.TeamID.String(),
		Name:   qt.Team.Name,
		Number: number,
	}, nil
}

func (s *QuizService) Buzz(ctx context.Context, eventID uuid.UUID, code string) (quiz_dto.BuzzResult, error) {
	const op = "service.QuizService.Buzz"
	log := s.log.With(slog.String("op", op))

	result := quiz_dto.BuzzResult{Position: -1}

	qt, err := s.teamRepository.GetByAccessCode(ctx, s.db, eventID, code)
	if err != nil {
		log.Error("failed get quiz team by code", logger.Err(err))
		return result, errs.ErrBadRequest.Wrap("Неверный код команды")
	}

	session, err := s.getSession(ctx, eventID)
	if err != nil {
		log.Error("failed get session", logger.Err(err))
		return result, errs.ErrInternalError
	}
	if session.CurrentRound != Round1 || !session.RoundActive || session.CurrentQuestion == 0 {
		return result, errs.ErrInvalidGameState.Wrap("Сейчас нет активного вопроса")
	}

	s.mu.Lock()
	t := s.timers[eventID]
	s.mu.Unlock()
	if t == nil || !t.started {
		// false start — record it, but the team may buzz again
		s.recordFalseStart(ctx, eventID, session.CurrentQuestion, qt.TeamID)
		return result, errs.ErrInvalidGameState.Wrap("Ложный старт: вопрос ещё не начался")
	}

	// idempotent re-buzz
	buzzes, err := s.buzzRepository.GetByQuestion(ctx, s.db, eventID, Round1, session.CurrentQuestion)
	if err != nil {
		log.Error("failed get buzzes", logger.Err(err))
		return result, errs.ErrInternalError
	}
	for i, b := range buzzes {
		if b.TeamID == qt.TeamID {
			if b.Status == quiz_buzz.StatusPending {
				result.Position = i + 1
				return result, nil
			}
			return result, errs.ErrInvalidGameState.Wrap("Команда уже отвечала на этот вопрос")
		}
	}

	pendingCount := 0
	var maxSeq int64 = 0
	for _, b := range buzzes {
		if b.Seq > maxSeq {
			maxSeq = b.Seq
		}
		if b.Status == quiz_buzz.StatusPending {
			pendingCount++
		}
	}

	buzz := quiz_buzz.QuizBuzz{
		EventID:        eventID,
		Round:          Round1,
		QuestionNumber: session.CurrentQuestion,
		TeamID:         qt.TeamID,
		Seq:            maxSeq + 1,
		Status:         quiz_buzz.StatusPending,
	}
	if err := s.buzzRepository.Save(ctx, s.db, buzz); err != nil {
		log.Error("failed save buzz", logger.Err(err))
		return result, errs.ErrInternalError
	}

	// pause the timer on the first pending buzz
	s.mu.Lock()
	if t.running {
		t.remaining = time.Until(t.deadline)
		t.running = false
	}
	s.mu.Unlock()

	if pendingCount+1 > MaxQueueRound1 {
		result.QueueFull = true
	}
	result.Position = pendingCount + 1

	if err := s.broadcastState(ctx, eventID); err != nil {
		return result, err
	}
	return result, nil
}

func (s *QuizService) recordFalseStart(ctx context.Context, eventID uuid.UUID, questionNumber int, teamID uuid.UUID) {
	buzzes, err := s.buzzRepository.GetByQuestion(ctx, s.db, eventID, Round1, questionNumber)
	if err != nil {
		return
	}
	var maxSeq int64 = 0
	for _, b := range buzzes {
		if b.Seq > maxSeq {
			maxSeq = b.Seq
		}
		if b.TeamID == teamID && b.Status != quiz_buzz.StatusFalseStart {
			return // already has a real entry
		}
	}
	_ = s.buzzRepository.Save(ctx, s.db, quiz_buzz.QuizBuzz{
		EventID:        eventID,
		Round:          Round1,
		QuestionNumber: questionNumber,
		TeamID:         teamID,
		Seq:            maxSeq + 1,
		Status:         quiz_buzz.StatusFalseStart,
	})
}

func (s *QuizService) ChooseOrder(ctx context.Context, eventID uuid.UUID, code string, order int) error {
	const op = "service.QuizService.ChooseOrder"
	log := s.log.With(slog.String("op", op))

	if order < quiz_team.OrderRed || order > quiz_team.OrderGreen {
		return errs.ErrBadRequest.Wrap("Некорректный орден")
	}

	qt, err := s.teamRepository.GetByAccessCode(ctx, s.db, eventID, code)
	if err != nil {
		log.Error("failed get quiz team by code", logger.Err(err))
		return errs.ErrBadRequest.Wrap("Неверный код команды")
	}
	if qt.Round3Order != quiz_team.OrderNone {
		return errs.ErrInvalidGameState.Wrap("Орден уже выбран и не может быть изменён")
	}

	session, err := s.getSession(ctx, eventID)
	if err != nil {
		log.Error("failed get session", logger.Err(err))
		return errs.ErrInternalError
	}
	if session.CurrentRound == Round3 && session.CurrentQuestion > 0 {
		return errs.ErrInvalidGameState.Wrap("Выбор ордена уже закрыт")
	}

	qt.Round3Order = order
	if err := s.teamRepository.Save(ctx, s.db, qt); err != nil {
		log.Error("failed save quiz team", logger.Err(err))
		return errs.ErrInternalError
	}
	return s.broadcastState(ctx, eventID)
}

// Grade evaluates the current question for a team. grade values:
// "yes"/"no" for round 1, "correct"/"partial"/"wrong" for round 2,
// "correct"/"wrong" for round 3.
func (s *QuizService) Grade(ctx context.Context, eventID uuid.UUID, code string, grade string) (quiz_dto.GradeResult, error) {
	const op = "service.QuizService.Grade"
	log := s.log.With(slog.String("op", op))

	result := quiz_dto.GradeResult{}

	qt, err := s.teamRepository.GetByAccessCode(ctx, s.db, eventID, code)
	if err != nil {
		log.Error("failed get quiz team by code", logger.Err(err))
		return result, errs.ErrBadRequest.Wrap("Неверный код команды")
	}

	session, err := s.getSession(ctx, eventID)
	if err != nil {
		log.Error("failed get session", logger.Err(err))
		return result, errs.ErrInternalError
	}
	if !session.RoundActive || session.CurrentQuestion == 0 {
		return result, errs.ErrInvalidGameState.Wrap("Нет активного вопроса")
	}

	switch session.CurrentRound {
	case Round1:
		return s.gradeRound1(ctx, eventID, session, qt, grade)
	case Round2:
		return s.gradeRound2(ctx, eventID, session, qt, grade)
	case Round3:
		return s.gradeRound3(ctx, eventID, session, qt, grade)
	default:
		return result, errs.ErrInvalidGameState.Wrap("Раунд не активен")
	}
}

func (s *QuizService) gradeRound1(ctx context.Context, eventID uuid.UUID, session quiz_session.QuizSession, qt quiz_team.QuizTeam, grade string) (quiz_dto.GradeResult, error) {
	const op = "service.QuizService.gradeRound1"
	log := s.log.With(slog.String("op", op))

	result := quiz_dto.GradeResult{}

	if grade != "yes" && grade != "no" {
		return result, errs.ErrBadRequest.Wrap("Некорректная оценка для первого раунда")
	}

	buzzes, err := s.buzzRepository.GetByQuestion(ctx, s.db, eventID, Round1, session.CurrentQuestion)
	if err != nil {
		log.Error("failed get buzzes", logger.Err(err))
		return result, errs.ErrInternalError
	}

	active := activeBuzz(buzzes)
	if active == nil || active.TeamID != qt.TeamID {
		return result, errs.ErrInvalidGameState.Wrap("Сейчас не очередь этой команды")
	}

	if grade == "yes" {
		active.Status = quiz_buzz.StatusCorrect
		if err := s.buzzRepository.Save(ctx, s.db, *active); err != nil {
			log.Error("failed save buzz", logger.Err(err))
			return result, errs.ErrInternalError
		}
		qt.Round1Score += 2
		if err := s.teamRepository.Save(ctx, s.db, qt); err != nil {
			log.Error("failed save team", logger.Err(err))
			return result, errs.ErrInternalError
		}
		result.PointsAwarded = 2

		// question concludes: stop timer, reveal answer
		s.mu.Lock()
		if t, ok := s.timers[eventID]; ok {
			t.remaining = 0
			t.running = false
			t.started = true
		}
		s.mu.Unlock()
		session.AnswerRevealed = true
		if err := s.sessionRepository.Save(ctx, s.db, session); err != nil {
			log.Error("failed save session", logger.Err(err))
			return result, errs.ErrInternalError
		}
	} else {
		active.Status = quiz_buzz.StatusWrong
		if err := s.buzzRepository.Save(ctx, s.db, *active); err != nil {
			log.Error("failed save buzz", logger.Err(err))
			return result, errs.ErrInternalError
		}

		if attemptsUsed(buzzes)+1 >= MaxAttemptsRound1 {
			// all attempts spent — question closes (host reveals manually)
			s.mu.Lock()
			if t, ok := s.timers[eventID]; ok {
				t.remaining = 0
				t.running = false
			}
			s.mu.Unlock()
		} else if next := activeBuzz(refreshBuzz(buzzes, *active)); next != nil {
			// turn passes to the next queued team; timer stays paused
		} else {
			// nobody left in queue — resume the timer for remaining teams
			s.mu.Lock()
			if t, ok := s.timers[eventID]; ok && t.started && !t.running && t.remaining > 0 {
				t.deadline = time.Now().Add(t.remaining)
				t.running = true
			}
			s.mu.Unlock()
		}
	}

	if err := s.broadcastState(ctx, eventID); err != nil {
		return result, err
	}
	return result, nil
}

// refreshBuzz returns the slice with the updated buzz swapped in.
func refreshBuzz(buzzes []quiz_buzz.QuizBuzz, updated quiz_buzz.QuizBuzz) []quiz_buzz.QuizBuzz {
	for i := range buzzes {
		if buzzes[i].TeamID == updated.TeamID {
			buzzes[i] = updated
		}
	}
	return buzzes
}

func (s *QuizService) gradeRound2(ctx context.Context, eventID uuid.UUID, session quiz_session.QuizSession, qt quiz_team.QuizTeam, grade string) (quiz_dto.GradeResult, error) {
	const op = "service.QuizService.gradeRound2"
	log := s.log.With(slog.String("op", op))

	result := quiz_dto.GradeResult{}

	var gradeValue int
	switch grade {
	case "correct":
		gradeValue = quiz_answer.GradeCorrect
	case "partial":
		gradeValue = quiz_answer.GradePartial
	case "wrong":
		gradeValue = quiz_answer.GradeWrong
	default:
		return result, errs.ErrBadRequest.Wrap("Некорректная оценка для второго раунда")
	}

	answer := quiz_answer.QuizAnswer{
		EventID:        eventID,
		Round:          Round2,
		QuestionNumber: session.CurrentQuestion,
		TeamID:         qt.TeamID,
		Grade:          gradeValue,
		Points:         quiz_answer.GradePoints(Round2, gradeValue),
	}
	if err := s.answerRepository.Upsert(ctx, s.db, answer); err != nil {
		log.Error("failed upsert answer", logger.Err(err))
		return result, errs.ErrInternalError
	}

	if err := s.recalculateRound2Score(ctx, eventID, qt); err != nil {
		log.Error("failed recalculate round2 score", logger.Err(err))
		return result, errs.ErrInternalError
	}

	result.PointsAwarded = answer.Points
	if err := s.broadcastState(ctx, eventID); err != nil {
		return result, err
	}
	return result, nil
}

func (s *QuizService) recalculateRound2Score(ctx context.Context, eventID uuid.UUID, qt quiz_team.QuizTeam) error {
	answers, err := s.answerRepository.GetByTeamAndRound(ctx, s.db, eventID, Round2, qt.TeamID)
	if err != nil {
		return err
	}
	total := 0
	for _, a := range answers {
		total += a.Points
	}
	qt.Round2Score = total
	return s.teamRepository.Save(ctx, s.db, qt)
}

func (s *QuizService) gradeRound3(ctx context.Context, eventID uuid.UUID, session quiz_session.QuizSession, qt quiz_team.QuizTeam, grade string) (quiz_dto.GradeResult, error) {
	const op = "service.QuizService.gradeRound3"
	log := s.log.With(slog.String("op", op))

	result := quiz_dto.GradeResult{}

	if qt.Round3Eliminated {
		return result, errs.ErrInvalidGameState.Wrap("Команда завершила участие в третьем раунде")
	}
	if qt.Round3Order == quiz_team.OrderNone {
		return result, errs.ErrInvalidGameState.Wrap("Орден не выбран")
	}

	var gradeValue int
	switch grade {
	case "correct":
		gradeValue = quiz_answer.GradeCorrect
	case "wrong":
		gradeValue = quiz_answer.GradeWrong
	default:
		return result, errs.ErrBadRequest.Wrap("Некорректная оценка для третьего раунда")
	}

	points := 0
	if gradeValue == quiz_answer.GradeCorrect {
		points = quiz_team.OrderPoints(qt.Round3Order)
	}

	answer := quiz_answer.QuizAnswer{
		EventID:        eventID,
		Round:          Round3,
		QuestionNumber: session.CurrentQuestion,
		TeamID:         qt.TeamID,
		Grade:          gradeValue,
		Points:         points,
	}
	if err := s.answerRepository.Upsert(ctx, s.db, answer); err != nil {
		log.Error("failed upsert answer", logger.Err(err))
		return result, errs.ErrInternalError
	}

	answers, err := s.answerRepository.GetByTeamAndRound(ctx, s.db, eventID, Round3, qt.TeamID)
	if err != nil {
		log.Error("failed get answers", logger.Err(err))
		return result, errs.ErrInternalError
	}

	total := 0
	wrongCount := 0
	for _, a := range answers {
		total += a.Points
		if a.Grade == quiz_answer.GradeWrong {
			wrongCount++
		}
	}
	qt.Round3Score = total
	qt.Round3WrongCount = wrongCount

	if wrongCount > quiz_team.OrderAllowedMistakes(qt.Round3Order) {
		qt.Round3Eliminated = true
		result.Eliminated = true
	}

	if err := s.teamRepository.Save(ctx, s.db, qt); err != nil {
		log.Error("failed save team", logger.Err(err))
		return result, errs.ErrInternalError
	}

	result.PointsAwarded = points
	if err := s.broadcastState(ctx, eventID); err != nil {
		return result, err
	}
	return result, nil
}

// ---------- state ----------

func (s *QuizService) GetPublicState(ctx context.Context, eventID uuid.UUID) (quiz_dto.PublicState, error) {
	const op = "service.QuizService.GetPublicState"
	log := s.log.With(slog.String("op", op))

	session, err := s.getSession(ctx, eventID)
	if err != nil {
		log.Error("failed get session", logger.Err(err))
		return quiz_dto.PublicState{}, errs.ErrInternalError
	}

	teams, err := s.teamRepository.GetByEventID(ctx, s.db, eventID)
	if err != nil {
		log.Error("failed get quiz teams", logger.Err(err))
		return quiz_dto.PublicState{}, errs.ErrInternalError
	}

	state := quiz_dto.PublicState{
		EventID:          eventID.String(),
		CurrentRound:     session.CurrentRound,
		CurrentQuestion:  session.CurrentQuestion,
		RoundActive:      session.RoundActive,
		AnswerRevealed:   session.AnswerRevealed,
		LeaderboardShown: session.LeaderboardShown,
		AttemptsMax:      MaxAttemptsRound1,
		Rounds:           RoundConfigs,
		BuzzQueue:        []quiz_dto.BuzzEntry{},
		Teams:            make([]quiz_dto.TeamState, 0, len(teams)),
	}

	cfg, cfgOk := roundConfig(session.CurrentRound)
	if cfgOk {
		t := s.getTimer(eventID, cfg)
		state.Timer = quiz_dto.TimerState{
			DurationSec: cfg.TimeSec,
			RemainingMs: t.remainingNow().Milliseconds(),
			Running:     t.running,
		}
	}

	if cfgOk && session.CurrentQuestion > 0 && session.CurrentQuestion <= cfg.QuestionCount {
		question, err := s.questionRepository.GetByNumber(ctx, s.db, eventID, session.CurrentRound, session.CurrentQuestion)
		if err == nil {
			qs := &quiz_dto.QuestionState{
				Number: question.QuestionNumber,
				Text:   question.Text,
			}
			if session.AnswerRevealed {
				qs.Answer = question.Answer
			}
			state.Question = qs
		}
	}

	if session.CurrentRound == Round1 && session.CurrentQuestion > 0 {
		buzzes, err := s.buzzRepository.GetByQuestion(ctx, s.db, eventID, Round1, session.CurrentQuestion)
		if err != nil {
			log.Error("failed get buzzes", logger.Err(err))
			return quiz_dto.PublicState{}, errs.ErrInternalError
		}
		teamNames := make(map[uuid.UUID]string, len(teams))
		for _, t := range teams {
			teamNames[t.TeamID] = t.Team.Name
		}
		active := activeBuzz(buzzes)
		pendingSeen := 0
		for _, b := range buzzes {
			if b.Status == quiz_buzz.StatusPending {
				pendingSeen++
			}
			state.BuzzQueue = append(state.BuzzQueue, quiz_dto.BuzzEntry{
				TeamID:   b.TeamID.String(),
				TeamName: teamNames[b.TeamID],
				Status:   b.Status,
				IsActive: active != nil && active.TeamID == b.TeamID,
			})
		}
		_ = pendingSeen
		state.AttemptsUsed = attemptsUsed(buzzes)
	}

	for i, t := range teams {
		total := t.Round1Score + t.Round2Score + t.Round3Score
		livesLeft := quiz_team.OrderAllowedMistakes(t.Round3Order) - t.Round3WrongCount
		if livesLeft < 0 {
			livesLeft = 0
		}
		state.Teams = append(state.Teams, quiz_dto.TeamState{
			TeamID:             t.TeamID.String(),
			Number:             i + 1,
			Name:               t.Team.Name,
			Round1Score:        t.Round1Score,
			Round2Score:        t.Round2Score,
			Round3Score:        t.Round3Score,
			TotalScore:         total,
			Round3Order:        t.Round3Order,
			Round3WrongCount:   t.Round3WrongCount,
			Round3LivesLeft:    livesLeft,
			Round3Eliminated:   t.Round3Eliminated,
		})
	}

	return state, nil
}

// broadcastState builds the public state and pushes it to all WS clients of the event.
func (s *QuizService) broadcastState(ctx context.Context, eventID uuid.UUID) error {
	state, err := s.GetPublicState(ctx, eventID)
	if err != nil {
		return err
	}
	s.hub.BroadcastToEvent(eventID.String(), connect_manager.Message{
		Action:  "quiz_state",
		Payload: state,
	})
	return nil
}
