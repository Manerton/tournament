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
	"main/internal/quiz/model/quiz_button"
	"main/internal/quiz/model/quiz_question"
	"main/internal/quiz/model/quiz_session"
	"main/internal/quiz/model/quiz_team"
	"main/internal/quiz/model/quiz_tiebreak_team"
	"main/internal/team/model/team"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	Round1 = 1
	Round2 = 2
	Round3 = 3
	// RoundReserve is the reserve question pool used by the tie-break (переигровка).
	RoundReserve = 4

	MaxAttemptsRound1 = 3

	// TiebreakTimeSec is the per-question timer of a tie-break (brain-ring mechanics).
	TiebreakTimeSec = 60

	// BuzzWindowDuration is the 1-second collection window opened by the first press:
	// all presses inside the window are recorded with their offsets, only the
	// fastest team gets the attempt.
	BuzzWindowDuration = time.Second
)

var RoundConfigs = []quiz_dto.RoundConfig{
	{Round: 1, Name: "Стрелы логики", QuestionCount: 20, TimeSec: 60},
	{Round: 2, Name: "Вектор удара", QuestionCount: 15, TimeSec: 60},
	{Round: 3, Name: "Формула клинка", QuestionCount: 10, TimeSec: 120},
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
	if t.remaining < 0 {
		return 0
	}
	return t.remaining
}

// buzzWindowState tracks the currently open 1-second press window of an event.
type buzzWindowState struct {
	question int // question number the window belongs to
	index    int // window number within the question
	start    time.Time
	deadline time.Time
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
	ResetUsedByEvent(ctx context.Context, db *gorm.DB, eventID uuid.UUID) error
}

type QuizTeamRepositoryInterface interface {
	GetByEventID(ctx context.Context, db *gorm.DB, eventID uuid.UUID) ([]quiz_team.QuizTeam, error)
	GetByTeamID(ctx context.Context, db *gorm.DB, eventID uuid.UUID, teamID uuid.UUID) (quiz_team.QuizTeam, error)
	GetByAccessCode(ctx context.Context, db *gorm.DB, eventID uuid.UUID, code string) (quiz_team.QuizTeam, error)
	Save(ctx context.Context, db *gorm.DB, team quiz_team.QuizTeam) error
	ResetScoresByEvent(ctx context.Context, db *gorm.DB, eventID uuid.UUID) error
}

type QuizBuzzRepositoryInterface interface {
	GetByQuestion(ctx context.Context, db *gorm.DB, eventID uuid.UUID, round int, questionNumber int) ([]quiz_buzz.QuizBuzz, error)
	GetByTeamAndQuestion(ctx context.Context, db *gorm.DB, eventID uuid.UUID, round int, questionNumber int, teamID uuid.UUID) (quiz_buzz.QuizBuzz, error)
	Save(ctx context.Context, db *gorm.DB, buzz quiz_buzz.QuizBuzz) error
	DeleteByQuestion(ctx context.Context, db *gorm.DB, eventID uuid.UUID, round int, questionNumber int) error
	DeleteByEvent(ctx context.Context, db *gorm.DB, eventID uuid.UUID) error
}

type QuizAnswerRepositoryInterface interface {
	GetByTeamAndRound(ctx context.Context, db *gorm.DB, eventID uuid.UUID, round int, teamID uuid.UUID) ([]quiz_answer.QuizAnswer, error)
	Upsert(ctx context.Context, db *gorm.DB, answer quiz_answer.QuizAnswer) error
	DeleteByQuestion(ctx context.Context, db *gorm.DB, eventID uuid.UUID, round int, questionNumber int) error
	DeleteByEvent(ctx context.Context, db *gorm.DB, eventID uuid.UUID) error
}

type TeamRepositoryInterface interface {
	GetAcceptTeamsByEventId(ctx context.Context, db *gorm.DB, eventID uuid.UUID) ([]team.Team, error)
}

type QuizButtonRepositoryInterface interface {
	GetByID(ctx context.Context, db *gorm.DB, id string) (quiz_button.QuizButton, error)
	Upsert(ctx context.Context, db *gorm.DB, button quiz_button.QuizButton) error
	Delete(ctx context.Context, db *gorm.DB, id string) error
	ListByEvent(ctx context.Context, db *gorm.DB, eventID uuid.UUID) ([]quiz_button.QuizButton, error)
}

type QuizTiebreakTeamRepositoryInterface interface {
	GetByEvent(ctx context.Context, db *gorm.DB, eventID uuid.UUID) ([]quiz_tiebreak_team.QuizTiebreakTeam, error)
	Save(ctx context.Context, db *gorm.DB, row quiz_tiebreak_team.QuizTiebreakTeam) error
	DeleteByEvent(ctx context.Context, db *gorm.DB, eventID uuid.UUID) error
}

type QuizService struct {
	log *slog.Logger
	db  *gorm.DB

	sessionRepository   QuizSessionRepositoryInterface
	questionRepository  QuizQuestionRepositoryInterface
	teamRepository      QuizTeamRepositoryInterface
	buzzRepository      QuizBuzzRepositoryInterface
	answerRepository    QuizAnswerRepositoryInterface
	eventTeamRepository TeamRepositoryInterface
	buttonRepository    QuizButtonRepositoryInterface
	tiebreakRepository  QuizTiebreakTeamRepositoryInterface

	hub *connect_manager.Hub

	mu          sync.Mutex
	timers      map[uuid.UUID]*timerRuntime
	buzzWindows map[uuid.UUID]*buzzWindowState
}

func New(log *slog.Logger, db *gorm.DB,
	sessionRepository QuizSessionRepositoryInterface,
	questionRepository QuizQuestionRepositoryInterface,
	teamRepository QuizTeamRepositoryInterface,
	buzzRepository QuizBuzzRepositoryInterface,
	answerRepository QuizAnswerRepositoryInterface,
	eventTeamRepository TeamRepositoryInterface,
	buttonRepository QuizButtonRepositoryInterface,
	tiebreakRepository QuizTiebreakTeamRepositoryInterface,
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
		buttonRepository:    buttonRepository,
		tiebreakRepository:  tiebreakRepository,
		hub:                 hub,
		timers:              make(map[uuid.UUID]*timerRuntime),
		buzzWindows:         make(map[uuid.UUID]*buzzWindowState),
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
	for i := range buzzes {
		if buzzes[i].Status == quiz_buzz.StatusPending {
			return &buzzes[i]
		}
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

// teamAttempted reports whether the team already got an attempt (graded) for this question.
func teamAttempted(buzzes []quiz_buzz.QuizBuzz, teamID uuid.UUID) bool {
	for _, b := range buzzes {
		if b.TeamID == teamID && (b.Status == quiz_buzz.StatusCorrect || b.Status == quiz_buzz.StatusWrong) {
			return true
		}
	}
	return false
}

func maxWindowIdx(buzzes []quiz_buzz.QuizBuzz) int {
	max := -1
	for _, b := range buzzes {
		if b.WindowIdx > max {
			max = b.WindowIdx
		}
	}
	return max
}

// pressesInWindow counts real presses (false starts excluded) of a window.
func pressesInWindow(buzzes []quiz_buzz.QuizBuzz, windowIdx int) int {
	n := 0
	for _, b := range buzzes {
		if b.WindowIdx == windowIdx && b.Status != quiz_buzz.StatusFalseStart {
			n++
		}
	}
	return n
}

// windowPlace returns the team's place in its window race (1 = fastest).
func windowPlace(buzzes []quiz_buzz.QuizBuzz, buzz quiz_buzz.QuizBuzz) int {
	place := 1
	for _, b := range buzzes {
		if b.WindowIdx == buzz.WindowIdx && b.OffsetMs < buzz.OffsetMs {
			place++
		}
	}
	return place
}

func nextSeq(buzzes []quiz_buzz.QuizBuzz) int64 {
	var maxSeq int64 = 0
	for _, b := range buzzes {
		if b.Seq > maxSeq {
			maxSeq = b.Seq
		}
	}
	return maxSeq + 1
}

func (s *QuizService) clearBuzzWindow(eventID uuid.UUID) {
	s.mu.Lock()
	delete(s.buzzWindows, eventID)
	s.mu.Unlock()
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
			Round1Score:      t.Round1Score + t.Round1Adjust,
			Round2Score:      t.Round2Score + t.Round2Adjust,
			Round3Score:      t.Round3Score + t.Round3Adjust,
			Round1Adjust:     t.Round1Adjust,
			Round2Adjust:     t.Round2Adjust,
			Round3Adjust:     t.Round3Adjust,
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

func (s *QuizService) UpsertQuestion(ctx context.Context, eventID uuid.UUID, round int, number int, text string, answer string, explanation string, criteria string) error {
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
		Explanation:    explanation,
		Criteria:       criteria,
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
	// the projector auto-shows the new round's rules slide at question 0
	session.RulesView = "none"

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
	session.RulesView = "none"
	if err := s.sessionRepository.Save(ctx, s.db, session); err != nil {
		log.Error("failed save session", logger.Err(err))
		return errs.ErrInternalError
	}

	// clear stale buzzes for this question
	if err := s.buzzRepository.DeleteByQuestion(ctx, s.db, eventID, session.CurrentRound, session.CurrentQuestion); err != nil {
		log.Error("failed clear buzzes", logger.Err(err))
		return errs.ErrInternalError
	}
	s.clearBuzzWindow(eventID)

	s.mu.Lock()
	t := s.timers[eventID]
	if t == nil {
		t = &timerRuntime{}
		s.timers[eventID] = t
	}
	t.duration = time.Duration(cfg.TimeSec) * time.Second
	t.remaining = t.duration
	t.deadline = time.Time{}
	t.running = false
	t.started = false
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

// ResumeTimer starts the timer of the current question or resumes it from a pause.
func (s *QuizService) ResumeTimer(ctx context.Context, eventID uuid.UUID) error {
	const op = "service.QuizService.ResumeTimer"
	log := s.log.With(slog.String("op", op))

	session, err := s.getSession(ctx, eventID)
	if err != nil {
		log.Error("failed get session", logger.Err(err))
		return errs.ErrInternalError
	}
	questionActive := (session.TiebreakActive && session.TiebreakQuestion > 0) ||
		(!session.TiebreakActive && session.RoundActive && session.CurrentQuestion > 0)
	if !questionActive {
		return errs.ErrInvalidGameState.Wrap("Нет активного вопроса")
	}

	s.mu.Lock()
	if t, ok := s.timers[eventID]; ok && !t.running && t.remaining > 0 {
		t.deadline = time.Now().Add(t.remaining)
		t.running = true
		t.started = true
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
	s.clearBuzzWindow(eventID)
	return s.broadcastState(ctx, eventID)
}

// CancelQuestion annuls the current question: all its results are cancelled
// for every team (points, mistakes, eliminations) and the question is closed.
func (s *QuizService) CancelQuestion(ctx context.Context, eventID uuid.UUID) error {
	const op = "service.QuizService.CancelQuestion"
	log := s.log.With(slog.String("op", op))

	session, err := s.getSession(ctx, eventID)
	if err != nil {
		log.Error("failed get session", logger.Err(err))
		return errs.ErrInternalError
	}
	if session.TiebreakActive {
		return errs.ErrInvalidGameState.Wrap("Вопрос переигровки отменить нельзя")
	}
	if !session.RoundActive || session.CurrentQuestion == 0 {
		return errs.ErrInvalidGameState.Wrap("Нет активного вопроса")
	}
	round, questionNumber := session.CurrentRound, session.CurrentQuestion

	question, qErr := s.questionRepository.GetByNumber(ctx, s.db, eventID, round, questionNumber)
	if qErr == nil && question.Cancelled {
		return errs.ErrInvalidGameState.Wrap("Вопрос уже отменён")
	}

	switch round {
	case Round1:
		// revoke the +2 of every correct answer, then wipe the buzz history
		buzzes, err := s.buzzRepository.GetByQuestion(ctx, s.db, eventID, Round1, questionNumber)
		if err != nil {
			log.Error("failed get buzzes", logger.Err(err))
			return errs.ErrInternalError
		}
		for _, b := range buzzes {
			if b.Status != quiz_buzz.StatusCorrect {
				continue
			}
			qt, err := s.teamRepository.GetByTeamID(ctx, s.db, eventID, b.TeamID)
			if err != nil {
				log.Error("failed get quiz team", logger.Err(err))
				return errs.ErrInternalError
			}
			qt.Round1Score -= 2
			if err := s.teamRepository.Save(ctx, s.db, qt); err != nil {
				log.Error("failed save team", logger.Err(err))
				return errs.ErrInternalError
			}
		}
		if err := s.buzzRepository.DeleteByQuestion(ctx, s.db, eventID, Round1, questionNumber); err != nil {
			log.Error("failed delete buzzes", logger.Err(err))
			return errs.ErrInternalError
		}
	case Round2, Round3:
		// scores, mistakes and eliminations are derived from the answer records:
		// dropping the question's answers and recalculating restores every team
		if err := s.answerRepository.DeleteByQuestion(ctx, s.db, eventID, round, questionNumber); err != nil {
			log.Error("failed delete answers", logger.Err(err))
			return errs.ErrInternalError
		}
		teams, err := s.teamRepository.GetByEventID(ctx, s.db, eventID)
		if err != nil {
			log.Error("failed get quiz teams", logger.Err(err))
			return errs.ErrInternalError
		}
		for _, qt := range teams {
			if round == Round2 {
				if err := s.recalculateRound2Score(ctx, eventID, qt); err != nil {
					log.Error("failed recalculate round2 score", logger.Err(err))
					return errs.ErrInternalError
				}
				continue
			}
			answers, err := s.answerRepository.GetByTeamAndRound(ctx, s.db, eventID, Round3, qt.TeamID)
			if err != nil {
				log.Error("failed get answers", logger.Err(err))
				return errs.ErrInternalError
			}
			total := 0
			wrongCount := 0
			for _, a := range answers {
				total += a.Points
				if a.Grade == quiz_answer.GradeWrong {
					wrongCount++
				}
			}
			qt.Round3Score = total + qt.Round3Adjust
			qt.Round3WrongCount = wrongCount
			qt.Round3Eliminated = wrongCount > qt.TotalAllowedMistakes()
			if err := s.teamRepository.Save(ctx, s.db, qt); err != nil {
				log.Error("failed save team", logger.Err(err))
				return errs.ErrInternalError
			}
		}
	default:
		return errs.ErrInvalidGameState.Wrap("Раунд не активен")
	}

	// mark the bank question as annulled (if it exists in the bank)
	if qErr == nil {
		question.Cancelled = true
		if err := s.questionRepository.Upsert(ctx, s.db, question); err != nil {
			log.Error("failed mark question cancelled", logger.Err(err))
			return errs.ErrInternalError
		}
	}

	// the cancelled question is over: stop the timer, drop the buzz window
	s.mu.Lock()
	if t, ok := s.timers[eventID]; ok {
		t.remaining = 0
		t.running = false
		t.started = true
	}
	s.mu.Unlock()
	s.clearBuzzWindow(eventID)

	log.Info("question cancelled",
		slog.String("event_id", eventID.String()),
		slog.Int("round", round),
		slog.Int("question", questionNumber))
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
	if show {
		// full-screen modes are mutually exclusive
		session.RulesView = "none"
	}
	if err := s.sessionRepository.Save(ctx, s.db, session); err != nil {
		log.Error("failed save session", logger.Err(err))
		return errs.ErrInternalError
	}
	return s.broadcastState(ctx, eventID)
}

// ShowRules switches the projector rules slide: "none", "overview", "1", "2" or "3".
func (s *QuizService) ShowRules(ctx context.Context, eventID uuid.UUID, view string) error {
	const op = "service.QuizService.ShowRules"
	log := s.log.With(slog.String("op", op))

	switch view {
	case "none", "overview", "1", "2", "3":
	default:
		return errs.ErrBadRequest.Wrap("Некорректный режим правил")
	}

	session, err := s.getSession(ctx, eventID)
	if err != nil {
		log.Error("failed get session", logger.Err(err))
		return errs.ErrInternalError
	}
	session.RulesView = view
	if view != "none" {
		// full-screen modes are mutually exclusive
		session.LeaderboardShown = false
	}
	if err := s.sessionRepository.Save(ctx, s.db, session); err != nil {
		log.Error("failed save session", logger.Err(err))
		return errs.ErrInternalError
	}
	return s.broadcastState(ctx, eventID)
}

// normalizeRulesView guards against an empty value from legacy rows.
func normalizeRulesView(view string) string {
	if view == "" {
		return "none"
	}
	return view
}

// GrantTurn manually gives the answering turn to a team (human-error correction).
// Works in round 1 and, while a tie-break is active, in the tie-break.
func (s *QuizService) GrantTurn(ctx context.Context, eventID uuid.UUID, teamID uuid.UUID) error {
	const op = "service.QuizService.GrantTurn"
	log := s.log.With(slog.String("op", op))

	session, err := s.getSession(ctx, eventID)
	if err != nil {
		log.Error("failed get session", logger.Err(err))
		return errs.ErrInternalError
	}

	round, question := session.CurrentRound, session.CurrentQuestion
	if session.TiebreakActive {
		round, question = RoundReserve, session.TiebreakQuestion
	} else if session.CurrentRound != Round1 {
		return errs.ErrInvalidGameState.Wrap("Сейчас нет активного вопроса первого раунда")
	}
	if question == 0 {
		return errs.ErrInvalidGameState.Wrap("Сейчас нет активного вопроса")
	}

	s.mu.Lock()
	t, hasTimer := s.timers[eventID]
	s.mu.Unlock()
	if !hasTimer || t.remainingNow() <= 0 {
		// time is up — no answers can be accepted anymore
		return errs.ErrInvalidGameState.Wrap("Время на вопрос вышло")
	}

	buzzes, err := s.buzzRepository.GetByQuestion(ctx, s.db, eventID, round, question)
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
			Round:          round,
			QuestionNumber: question,
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
	if t.running {
		t.remaining = time.Until(t.deadline)
		if t.remaining < 0 {
			t.remaining = 0
		}
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
	if session.CurrentQuestion == 0 && !session.TiebreakActive {
		return errs.ErrInvalidGameState.Wrap("Нет активного вопроса")
	}
	if session.TiebreakActive && session.TiebreakQuestion == 0 {
		return errs.ErrInvalidGameState.Wrap("Нет активного вопроса")
	}

	round, question := session.CurrentRound, session.CurrentQuestion
	if session.TiebreakActive {
		round, question = RoundReserve, session.TiebreakQuestion
	}

	if err := s.buzzRepository.DeleteByQuestion(ctx, s.db, eventID, round, question); err != nil {
		log.Error("failed clear buzzes", logger.Err(err))
		return errs.ErrInternalError
	}
	s.clearBuzzWindow(eventID)

	s.mu.Lock()
	if t, ok := s.timers[eventID]; ok && t.started && !t.running {
		t.deadline = time.Now().Add(t.remaining)
		t.running = true
	}
	s.mu.Unlock()

	return s.broadcastState(ctx, eventID)
}

// ResetTournament wipes all quiz progress of the event: buzzes, answers,
// tiebreak data and every team's scores (including manual adjustments and
// round-3 order state), returns the session to the initial state, and marks
// reserve-pool questions as unused again.
// Teams, access codes and questions are kept.
func (s *QuizService) ResetTournament(ctx context.Context, eventID uuid.UUID) error {
	const op = "service.QuizService.ResetTournament"
	log := s.log.With(slog.String("op", op))

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.buzzRepository.DeleteByEvent(ctx, tx, eventID); err != nil {
			return err
		}
		if err := s.answerRepository.DeleteByEvent(ctx, tx, eventID); err != nil {
			return err
		}
		if err := s.tiebreakRepository.DeleteByEvent(ctx, tx, eventID); err != nil {
			return err
		}
		if err := s.teamRepository.ResetScoresByEvent(ctx, tx, eventID); err != nil {
			return err
		}
		if err := s.questionRepository.ResetUsedByEvent(ctx, tx, eventID); err != nil {
			return err
		}

		session, err := s.sessionRepository.GetOrCreate(ctx, tx, eventID)
		if err != nil {
			return err
		}
		session.CurrentRound = 0
		session.CurrentQuestion = 0
		session.AnswerRevealed = false
		session.RoundActive = false
		session.TiebreakActive = false
		session.TiebreakQuestion = 0
		session.RulesView = "none"
		return s.sessionRepository.Save(ctx, tx, session)
	})
	if err != nil {
		log.Error("failed reset tournament", logger.Err(err))
		return errs.ErrInternalError
	}

	s.mu.Lock()
	delete(s.timers, eventID)
	delete(s.buzzWindows, eventID)
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

// resolveBuzzTeam identifies the pressing team: either by ESP32 button id
// (preconfigured mapping) or by herald access code (manual press).
func (s *QuizService) resolveBuzzTeam(ctx context.Context, eventID uuid.UUID, code string, buttonID string) (quiz_team.QuizTeam, error) {
	if buttonID != "" {
		button, err := s.buttonRepository.GetByID(ctx, s.db, buttonID)
		if err != nil || button.QuizTeam.EventID != eventID {
			return quiz_team.QuizTeam{}, errs.ErrBadRequest.Wrap("Неизвестная кнопка")
		}
		return button.QuizTeam, nil
	}
	qt, err := s.teamRepository.GetByAccessCode(ctx, s.db, eventID, code)
	if err != nil {
		return quiz_team.QuizTeam{}, errs.ErrBadRequest.Wrap("Неверный код команды")
	}
	return qt, nil
}

// Buzz handles a press (ESP32 button or manual herald buzz) in round 1 and,
// while a tie-break is active, in the tie-break (with the whitelist lock).
// The first press opens a 1-second window: the fastest team gets the attempt,
// other presses inside the window are recorded with their offsets for display only.
// After a failed attempt the next press opens a fresh window; a team that already
// answered cannot press again for the same question.
func (s *QuizService) Buzz(ctx context.Context, eventID uuid.UUID, code string, buttonID string) (quiz_dto.BuzzResult, error) {
	const op = "service.QuizService.Buzz"
	log := s.log.With(slog.String("op", op))

	result := quiz_dto.BuzzResult{Position: -1}

	qt, err := s.resolveBuzzTeam(ctx, eventID, code, buttonID)
	if err != nil {
		log.Error("failed resolve buzz team", logger.Err(err))
		return result, err
	}

	session, err := s.getSession(ctx, eventID)
	if err != nil {
		log.Error("failed get session", logger.Err(err))
		return result, errs.ErrInternalError
	}

	if session.TiebreakActive {
		return s.buzzTiebreak(ctx, eventID, session, qt)
	}

	if session.CurrentRound != Round1 || !session.RoundActive || session.CurrentQuestion == 0 {
		return result, errs.ErrInvalidGameState.Wrap("Сейчас нет активного вопроса")
	}

	return s.buzzWindowed(ctx, eventID, qt, Round1, session.CurrentQuestion, MaxAttemptsRound1)
}

// buzzWindowed runs the shared buzzer mechanics of round 1 and the tie-break:
// the first press opens a 1-second window, the fastest team gets the attempt,
// other presses inside the window are recorded with their offsets for display.
// maxAttempts caps the graded attempts per question.
func (s *QuizService) buzzWindowed(ctx context.Context, eventID uuid.UUID, qt quiz_team.QuizTeam, round int, questionNumber int, maxAttempts int) (quiz_dto.BuzzResult, error) {
	const op = "service.QuizService.buzzWindowed"
	log := s.log.With(slog.String("op", op))

	result := quiz_dto.BuzzResult{Position: -1}

	s.mu.Lock()
	t := s.timers[eventID]
	s.mu.Unlock()
	if t == nil || !t.started {
		// false start — record it, but the team may buzz again
		s.recordFalseStart(ctx, eventID, round, questionNumber, qt.TeamID)
		return result, errs.ErrInvalidGameState.Wrap("Ложный старт: вопрос ещё не начался")
	}
	if t.remainingNow() <= 0 {
		// time is up — the question is closed, no more presses are accepted
		s.clearBuzzWindow(eventID)
		return result, errs.ErrInvalidGameState.Wrap("Время на вопрос вышло")
	}

	buzzes, err := s.buzzRepository.GetByQuestion(ctx, s.db, eventID, round, questionNumber)
	if err != nil {
		log.Error("failed get buzzes", logger.Err(err))
		return result, errs.ErrInternalError
	}

	// idempotent re-buzz while this team holds the attempt
	for _, b := range buzzes {
		if b.TeamID == qt.TeamID && b.Status == quiz_buzz.StatusPending {
			result.Position = windowPlace(buzzes, b)
			result.OffsetMs = b.OffsetMs
			return result, nil
		}
	}
	if teamAttempted(buzzes, qt.TeamID) {
		return result, errs.ErrInvalidGameState.Wrap("Команда уже отвечала на этот вопрос")
	}

	now := time.Now()

	s.mu.Lock()
	w := s.buzzWindows[eventID]
	if w != nil && w.question != questionNumber {
		delete(s.buzzWindows, eventID)
		w = nil
	}

	if w == nil {
		// first press of a new window — this team gets the attempt
		if attemptsUsed(buzzes) >= maxAttempts {
			s.mu.Unlock()
			return result, errs.ErrInvalidGameState.Wrap("Попытки на этот вопрос закончились")
		}
		w = &buzzWindowState{
			question: questionNumber,
			index:    maxWindowIdx(buzzes) + 1,
			start:    now,
			deadline: now.Add(BuzzWindowDuration),
		}
		s.buzzWindows[eventID] = w
		s.mu.Unlock()

		buzz := quiz_buzz.QuizBuzz{
			EventID:        eventID,
			Round:          round,
			QuestionNumber: questionNumber,
			TeamID:         qt.TeamID,
			Seq:            nextSeq(buzzes),
			Status:         quiz_buzz.StatusPending,
			WindowIdx:      w.index,
			OffsetMs:       0,
		}
		if err := s.buzzRepository.Save(ctx, s.db, buzz); err != nil {
			log.Error("failed save buzz", logger.Err(err))
			return result, errs.ErrInternalError
		}

		// pause the timer while the team answers
		s.mu.Lock()
		if t.running {
			t.remaining = time.Until(t.deadline)
			if t.remaining < 0 {
				t.remaining = 0
			}
			t.running = false
		}
		s.mu.Unlock()

		result.Position = 1
		result.OffsetMs = 0
		if err := s.broadcastState(ctx, eventID); err != nil {
			return result, err
		}
		return result, nil
	}
	window := *w
	s.mu.Unlock()

	// window already open — the press is recorded for the race display only
	if now.After(window.deadline) {
		return result, errs.ErrInvalidGameState.Wrap("Приём нажатий закрыт, отвечает другая команда")
	}

	offset := int(now.Sub(window.start).Milliseconds())
	buzz := quiz_buzz.QuizBuzz{
		EventID:        eventID,
		Round:          round,
		QuestionNumber: questionNumber,
		TeamID:         qt.TeamID,
		Seq:            nextSeq(buzzes),
		Status:         quiz_buzz.StatusOutpaced,
		WindowIdx:      window.index,
		OffsetMs:       offset,
	}
	if err := s.buzzRepository.Save(ctx, s.db, buzz); err != nil {
		log.Error("failed save buzz", logger.Err(err))
		return result, errs.ErrInternalError
	}

	result.Position = pressesInWindow(buzzes, window.index) + 1
	result.OffsetMs = offset
	if err := s.broadcastState(ctx, eventID); err != nil {
		return result, err
	}
	return result, nil
}

func (s *QuizService) recordFalseStart(ctx context.Context, eventID uuid.UUID, round int, questionNumber int, teamID uuid.UUID) {
	buzzes, err := s.buzzRepository.GetByQuestion(ctx, s.db, eventID, round, questionNumber)
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
		Round:          round,
		QuestionNumber: questionNumber,
		TeamID:         teamID,
		Seq:            maxSeq + 1,
		Status:         quiz_buzz.StatusFalseStart,
		WindowIdx:      -1, // false starts are not part of any window
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

// AdjustScore applies a manual score correction (delta) for one round of a team.
// Corrections accumulate in per-round adjust columns so automatic round
// recalculations (rounds 2/3 rebuild scores from answer records) never wipe them.
func (s *QuizService) AdjustScore(ctx context.Context, eventID uuid.UUID, teamID uuid.UUID, round int, delta int) error {
	const op = "service.QuizService.AdjustScore"
	log := s.log.With(slog.String("op", op))

	if round < Round1 || round > Round3 {
		return errs.ErrBadRequest.Wrap("Некорректный раунд")
	}

	qt, err := s.teamRepository.GetByTeamID(ctx, s.db, eventID, teamID)
	if err != nil {
		log.Error("failed get quiz team", logger.Err(err))
		return errs.ErrBadRequest.Wrap("Команда не найдена")
	}

	switch round {
	case Round1:
		qt.Round1Adjust += delta
	case Round2:
		qt.Round2Adjust += delta
	case Round3:
		qt.Round3Adjust += delta
	}
	if err := s.teamRepository.Save(ctx, s.db, qt); err != nil {
		log.Error("failed save quiz team", logger.Err(err))
		return errs.ErrInternalError
	}
	return s.broadcastState(ctx, eventID)
}

// AdminSetOrder sets (or resets, when order is OrderNone) a team's round-3 order,
// bypassing the herald restriction — for correcting herald mistakes.
func (s *QuizService) AdminSetOrder(ctx context.Context, eventID uuid.UUID, teamID uuid.UUID, order int) error {
	const op = "service.QuizService.AdminSetOrder"
	log := s.log.With(slog.String("op", op))

	if order < quiz_team.OrderNone || order > quiz_team.OrderGreen {
		return errs.ErrBadRequest.Wrap("Некорректный орден")
	}

	qt, err := s.teamRepository.GetByTeamID(ctx, s.db, eventID, teamID)
	if err != nil {
		log.Error("failed get quiz team", logger.Err(err))
		return errs.ErrBadRequest.Wrap("Команда не найдена")
	}

	qt.Round3Order = order
	if err := s.teamRepository.Save(ctx, s.db, qt); err != nil {
		log.Error("failed save quiz team", logger.Err(err))
		return errs.ErrInternalError
	}
	return s.broadcastState(ctx, eventID)
}

// RestoreTeam returns an eliminated round-3 team to the competition. Extra
// lives are granted to cover the team's accumulated mistakes, so the team
// continues without the right to another error: the next wrong grade
// eliminates it again.
func (s *QuizService) RestoreTeam(ctx context.Context, eventID uuid.UUID, teamID uuid.UUID) error {
	const op = "service.QuizService.RestoreTeam"
	log := s.log.With(slog.String("op", op))

	qt, err := s.teamRepository.GetByTeamID(ctx, s.db, eventID, teamID)
	if err != nil {
		log.Error("failed get quiz team", logger.Err(err))
		return errs.ErrBadRequest.Wrap("Команда не найдена")
	}
	if !qt.Round3Eliminated {
		return errs.ErrInvalidGameState.Wrap("Команда не выбыла из третьего тура")
	}

	if deficit := qt.Round3WrongCount - qt.TotalAllowedMistakes(); deficit > 0 {
		qt.Round3ExtraLives += deficit
	}
	qt.Round3Eliminated = false
	if err := s.teamRepository.Save(ctx, s.db, qt); err != nil {
		log.Error("failed save quiz team", logger.Err(err))
		return errs.ErrInternalError
	}

	log.Info("team restored to round 3",
		slog.String("event_id", eventID.String()),
		slog.String("team_id", teamID.String()),
		slog.Int("extra_lives", qt.Round3ExtraLives))
	return s.broadcastState(ctx, eventID)
}

// ---------- buttons (ESP32) ----------

func (s *QuizService) ListButtons(ctx context.Context, eventID uuid.UUID) ([]quiz_dto.AdminButton, error) {
	const op = "service.QuizService.ListButtons"
	log := s.log.With(slog.String("op", op))

	buttons, err := s.buttonRepository.ListByEvent(ctx, s.db, eventID)
	if err != nil {
		log.Error("failed list buttons", logger.Err(err))
		return nil, errs.ErrInternalError
	}

	result := make([]quiz_dto.AdminButton, 0, len(buttons))
	for _, b := range buttons {
		result = append(result, quiz_dto.AdminButton{
			ButtonID: b.ID,
			TeamID:   b.QuizTeam.TeamID.String(),
			TeamName: b.QuizTeam.Team.Name,
		})
	}
	return result, nil
}

// RegisterButton binds a physical button id to a quiz team (re-binding allowed).
func (s *QuizService) RegisterButton(ctx context.Context, eventID uuid.UUID, buttonID string, teamID uuid.UUID) error {
	const op = "service.QuizService.RegisterButton"
	log := s.log.With(slog.String("op", op))

	buttonID = strings.TrimSpace(buttonID)
	if buttonID == "" || len(buttonID) > 64 {
		return errs.ErrBadRequest.Wrap("Некорректный id кнопки")
	}

	qt, err := s.teamRepository.GetByTeamID(ctx, s.db, eventID, teamID)
	if err != nil {
		log.Error("failed get quiz team", logger.Err(err))
		return errs.ErrBadRequest.Wrap("Команда не найдена")
	}

	if err := s.buttonRepository.Upsert(ctx, s.db, quiz_button.QuizButton{
		ID:         buttonID,
		QuizTeamID: qt.ID,
	}); err != nil {
		log.Error("failed upsert button", logger.Err(err))
		return errs.ErrInternalError
	}
	return s.broadcastState(ctx, eventID)
}

func (s *QuizService) DeleteButton(ctx context.Context, buttonID string) error {
	const op = "service.QuizService.DeleteButton"
	log := s.log.With(slog.String("op", op))

	if err := s.buttonRepository.Delete(ctx, s.db, buttonID); err != nil {
		log.Error("failed delete button", logger.Err(err))
		return errs.ErrInternalError
	}
	return nil
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

	if session.TiebreakActive {
		return s.gradeTiebreak(ctx, eventID, session, qt, grade)
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
		s.clearBuzzWindow(eventID)
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

		// count attempts exactly as they stand after this grade
		// (the fetched slice is stale — it still has this buzz as pending)
		used := attemptsUsed(refreshBuzz(buzzes, *active))

		s.mu.Lock()
		t, hasTimer := s.timers[eventID]
		expired := !hasTimer || t.remainingNow() <= 0
		if used >= MaxAttemptsRound1 {
			// all attempts spent — question closes (host reveals manually)
			log.Info("round 1 question closed: attempts spent",
				slog.String("event_id", eventID.String()),
				slog.Int("question", session.CurrentQuestion),
				slog.Int("attempts_used", used))
			if hasTimer {
				t.remaining = 0
				t.running = false
			}
		} else if expired {
			// time ran out (e.g. while no one was answering) — question closes
			log.Info("round 1 question closed: time expired after wrong answer",
				slog.String("event_id", eventID.String()),
				slog.Int("question", session.CurrentQuestion),
				slog.Int("attempts_used", used))
			if hasTimer {
				t.remaining = 0
				t.running = false
			}
		} else if t.running {
			// the timer was resumed mid-answer (manual resume) — let it keep running
			log.Info("round 1 answer rejected: timer left running (resumed mid-answer)",
				slog.String("event_id", eventID.String()),
				slog.Int("question", session.CurrentQuestion),
				slog.Int("attempts_used", used))
		} else {
			// answer failed — resume the timer; the next press opens a fresh window
			t.deadline = time.Now().Add(t.remaining)
			t.running = true
			log.Info("round 1 answer rejected: timer resumed",
				slog.String("event_id", eventID.String()),
				slog.Int("question", session.CurrentQuestion),
				slog.Int("attempts_used", used),
				slog.Int64("remaining_ms", t.remaining.Milliseconds()))
		}
		s.mu.Unlock()
		s.clearBuzzWindow(eventID)
	}

	if err := s.broadcastState(ctx, eventID); err != nil {
		return result, err
	}
	return result, nil
}

// refreshBuzz returns the slice with the updated buzz swapped in, matched by id
// (a team may hold several entries in one question — e.g. outpaced + pending).
func refreshBuzz(buzzes []quiz_buzz.QuizBuzz, updated quiz_buzz.QuizBuzz) []quiz_buzz.QuizBuzz {
	for i := range buzzes {
		if buzzes[i].ID == updated.ID {
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

	if question, err := s.questionRepository.GetByNumber(ctx, s.db, eventID, Round2, session.CurrentQuestion); err == nil && question.Cancelled {
		return result, errs.ErrInvalidGameState.Wrap("Вопрос отменён — оценка не требуется")
	}

	// grading opens only after the question time is over (or the question is closed)
	s.mu.Lock()
	t, hasTimer := s.timers[eventID]
	timeLeft := hasTimer && t.remainingNow() > 0
	s.mu.Unlock()
	if timeLeft {
		return result, errs.ErrInvalidGameState.Wrap("Оценивание откроется после окончания времени")
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
	qt.Round2Score = total + qt.Round2Adjust
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
	if question, err := s.questionRepository.GetByNumber(ctx, s.db, eventID, Round3, session.CurrentQuestion); err == nil && question.Cancelled {
		return result, errs.ErrInvalidGameState.Wrap("Вопрос отменён — оценка не требуется")
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
	qt.Round3Score = total + qt.Round3Adjust
	qt.Round3WrongCount = wrongCount

	if wrongCount > qt.TotalAllowedMistakes() {
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

// ---------- tie-break (переигровка) ----------

// activeTiebreakTeams returns the group members that have not secured a place yet.
func activeTiebreakTeams(rows []quiz_tiebreak_team.QuizTiebreakTeam) []quiz_tiebreak_team.QuizTiebreakTeam {
	active := make([]quiz_tiebreak_team.QuizTiebreakTeam, 0, len(rows))
	for _, r := range rows {
		if !r.Won {
			active = append(active, r)
		}
	}
	return active
}

func findTiebreakTeam(rows []quiz_tiebreak_team.QuizTiebreakTeam, teamID uuid.UUID) *quiz_tiebreak_team.QuizTiebreakTeam {
	for i := range rows {
		if rows[i].TeamID == teamID {
			return &rows[i]
		}
	}
	return nil
}

// StartTiebreak opens a tie-break for the selected tied group (2+ teams).
// Only these teams may buzz until the tie-break ends.
func (s *QuizService) StartTiebreak(ctx context.Context, eventID uuid.UUID, teamIDs []uuid.UUID) error {
	const op = "service.QuizService.StartTiebreak"
	log := s.log.With(slog.String("op", op))

	unique := make(map[uuid.UUID]bool, len(teamIDs))
	for _, id := range teamIDs {
		unique[id] = true
	}
	if len(unique) < 2 {
		return errs.ErrBadRequest.Wrap("Для переигровки нужно выбрать минимум две команды")
	}

	session, err := s.getSession(ctx, eventID)
	if err != nil {
		log.Error("failed get session", logger.Err(err))
		return errs.ErrInternalError
	}
	if session.TiebreakActive {
		return errs.ErrInvalidGameState.Wrap("Переигровка уже идёт")
	}

	for id := range unique {
		if _, err := s.teamRepository.GetByTeamID(ctx, s.db, eventID, id); err != nil {
			return errs.ErrBadRequest.Wrap("Команда не найдена")
		}
	}
	if err := s.tiebreakRepository.DeleteByEvent(ctx, s.db, eventID); err != nil {
		log.Error("failed clear tiebreak group", logger.Err(err))
		return errs.ErrInternalError
	}
	for id := range unique {
		if err := s.tiebreakRepository.Save(ctx, s.db, quiz_tiebreak_team.QuizTiebreakTeam{
			EventID: eventID,
			TeamID:  id,
		}); err != nil {
			log.Error("failed save tiebreak team", logger.Err(err))
			return errs.ErrInternalError
		}
	}

	session.TiebreakActive = true
	session.TiebreakQuestion = 0
	session.AnswerRevealed = false
	session.LeaderboardShown = false
	// the projector auto-shows the tie-break rules slide at question 0
	session.RulesView = "none"
	if err := s.sessionRepository.Save(ctx, s.db, session); err != nil {
		log.Error("failed save session", logger.Err(err))
		return errs.ErrInternalError
	}

	s.mu.Lock()
	s.timers[eventID] = &timerRuntime{
		duration:  TiebreakTimeSec * time.Second,
		remaining: TiebreakTimeSec * time.Second,
	}
	s.mu.Unlock()

	return s.broadcastState(ctx, eventID)
}

// StartTiebreakQuestion shows the next unused reserve question. A shown question
// is marked used and never returns: the whole hall shares one projector.
func (s *QuizService) StartTiebreakQuestion(ctx context.Context, eventID uuid.UUID) error {
	const op = "service.QuizService.StartTiebreakQuestion"
	log := s.log.With(slog.String("op", op))

	session, err := s.getSession(ctx, eventID)
	if err != nil {
		log.Error("failed get session", logger.Err(err))
		return errs.ErrInternalError
	}
	if !session.TiebreakActive {
		return errs.ErrInvalidGameState.Wrap("Переигровка не начата")
	}

	questions, err := s.questionRepository.GetByEventAndRound(ctx, s.db, eventID, RoundReserve)
	if err != nil {
		log.Error("failed get reserve questions", logger.Err(err))
		return errs.ErrInternalError
	}
	var next *quiz_question.QuizQuestion
	for i := range questions {
		if !questions[i].Used {
			next = &questions[i]
			break
		}
	}
	if next == nil {
		return errs.ErrInvalidGameState.Wrap("Резервные вопросы закончились — используйте решение жюри")
	}

	next.Used = true
	if err := s.questionRepository.Upsert(ctx, s.db, *next); err != nil {
		log.Error("failed mark question used", logger.Err(err))
		return errs.ErrInternalError
	}

	session.TiebreakQuestion = next.QuestionNumber
	session.AnswerRevealed = false
	session.RulesView = "none"
	if err := s.sessionRepository.Save(ctx, s.db, session); err != nil {
		log.Error("failed save session", logger.Err(err))
		return errs.ErrInternalError
	}

	if err := s.buzzRepository.DeleteByQuestion(ctx, s.db, eventID, RoundReserve, next.QuestionNumber); err != nil {
		log.Error("failed clear buzzes", logger.Err(err))
		return errs.ErrInternalError
	}
	s.clearBuzzWindow(eventID)

	s.mu.Lock()
	s.timers[eventID] = &timerRuntime{
		duration:  TiebreakTimeSec * time.Second,
		remaining: TiebreakTimeSec * time.Second,
	}
	s.mu.Unlock()

	return s.broadcastState(ctx, eventID)
}

// EndTiebreak closes the tie-break mode and clears the tied group.
func (s *QuizService) EndTiebreak(ctx context.Context, eventID uuid.UUID) error {
	const op = "service.QuizService.EndTiebreak"
	log := s.log.With(slog.String("op", op))

	session, err := s.getSession(ctx, eventID)
	if err != nil {
		log.Error("failed get session", logger.Err(err))
		return errs.ErrInternalError
	}

	session.TiebreakActive = false
	session.TiebreakQuestion = 0
	session.AnswerRevealed = false
	if err := s.sessionRepository.Save(ctx, s.db, session); err != nil {
		log.Error("failed save session", logger.Err(err))
		return errs.ErrInternalError
	}

	if err := s.tiebreakRepository.DeleteByEvent(ctx, s.db, eventID); err != nil {
		log.Error("failed clear tiebreak group", logger.Err(err))
		return errs.ErrInternalError
	}

	s.mu.Lock()
	if t, ok := s.timers[eventID]; ok {
		t.remaining = 0
		t.running = false
		t.started = true
	}
	s.mu.Unlock()
	s.clearBuzzWindow(eventID)

	return s.broadcastState(ctx, eventID)
}

// JuryDecision is the fallback for when reserve questions run out (or the jury
// decides a place directly): the host manually sets a team's tiebreak score.
func (s *QuizService) JuryDecision(ctx context.Context, eventID uuid.UUID, teamID uuid.UUID, score int) error {
	const op = "service.QuizService.JuryDecision"
	log := s.log.With(slog.String("op", op))

	qt, err := s.teamRepository.GetByTeamID(ctx, s.db, eventID, teamID)
	if err != nil {
		log.Error("failed get quiz team", logger.Err(err))
		return errs.ErrBadRequest.Wrap("Команда не найдена")
	}

	qt.TiebreakScore = score
	if err := s.teamRepository.Save(ctx, s.db, qt); err != nil {
		log.Error("failed save quiz team", logger.Err(err))
		return errs.ErrInternalError
	}
	return s.broadcastState(ctx, eventID)
}

// buzzTiebreak applies the whitelist lock (only the tied group may buzz),
// then runs the shared buzzer mechanics. Attempts per question equal the
// number of teams still competing in the group.
func (s *QuizService) buzzTiebreak(ctx context.Context, eventID uuid.UUID, session quiz_session.QuizSession, qt quiz_team.QuizTeam) (quiz_dto.BuzzResult, error) {
	result := quiz_dto.BuzzResult{Position: -1}

	if session.TiebreakQuestion == 0 {
		return result, errs.ErrInvalidGameState.Wrap("Резервный вопрос ещё не начат")
	}

	rows, err := s.tiebreakRepository.GetByEvent(ctx, s.db, eventID)
	if err != nil {
		return result, errs.ErrInternalError
	}
	member := findTiebreakTeam(rows, qt.TeamID)
	if member == nil || member.Won {
		return result, errs.ErrInvalidGameState.Wrap("Команда не участвует в переигровке")
	}

	return s.buzzWindowed(ctx, eventID, qt, RoundReserve, session.TiebreakQuestion, len(activeTiebreakTeams(rows)))
}

// gradeTiebreak grades the current reserve question ("yes"/"no", brain-ring style).
// A correct answer wins the step: the team secures the highest place still
// contested in the tied group and leaves it. Round scores and the total are
// never modified — only the tiebreak_score sort key.
func (s *QuizService) gradeTiebreak(ctx context.Context, eventID uuid.UUID, session quiz_session.QuizSession, qt quiz_team.QuizTeam, grade string) (quiz_dto.GradeResult, error) {
	const op = "service.QuizService.gradeTiebreak"
	log := s.log.With(slog.String("op", op))

	result := quiz_dto.GradeResult{}

	if grade != "yes" && grade != "no" {
		return result, errs.ErrBadRequest.Wrap("Некорректная оценка для переигровки")
	}
	if session.TiebreakQuestion == 0 {
		return result, errs.ErrInvalidGameState.Wrap("Резервный вопрос ещё не начат")
	}

	rows, err := s.tiebreakRepository.GetByEvent(ctx, s.db, eventID)
	if err != nil {
		log.Error("failed get tiebreak group", logger.Err(err))
		return result, errs.ErrInternalError
	}
	member := findTiebreakTeam(rows, qt.TeamID)
	if member == nil || member.Won {
		return result, errs.ErrInvalidGameState.Wrap("Команда не участвует в переигровке")
	}
	active := activeTiebreakTeams(rows)

	buzzes, err := s.buzzRepository.GetByQuestion(ctx, s.db, eventID, RoundReserve, session.TiebreakQuestion)
	if err != nil {
		log.Error("failed get buzzes", logger.Err(err))
		return result, errs.ErrInternalError
	}

	activeB := activeBuzz(buzzes)
	if activeB == nil || activeB.TeamID != qt.TeamID {
		return result, errs.ErrInvalidGameState.Wrap("Сейчас не очередь этой команды")
	}

	if grade == "yes" {
		activeB.Status = quiz_buzz.StatusCorrect
		if err := s.buzzRepository.Save(ctx, s.db, *activeB); err != nil {
			log.Error("failed save buzz", logger.Err(err))
			return result, errs.ErrInternalError
		}

		member.Won = true
		if err := s.tiebreakRepository.Save(ctx, s.db, *member); err != nil {
			log.Error("failed save tiebreak team", logger.Err(err))
			return result, errs.ErrInternalError
		}
		// the step winner takes the highest place still contested in the group
		qt.TiebreakScore = len(active)
		if err := s.teamRepository.Save(ctx, s.db, qt); err != nil {
			log.Error("failed save team", logger.Err(err))
			return result, errs.ErrInternalError
		}

		// question concludes: stop timer, reveal answer
		s.mu.Lock()
		if t, ok := s.timers[eventID]; ok {
			t.remaining = 0
			t.running = false
			t.started = true
		}
		s.mu.Unlock()
		s.clearBuzzWindow(eventID)
		session.AnswerRevealed = true
		if err := s.sessionRepository.Save(ctx, s.db, session); err != nil {
			log.Error("failed save session", logger.Err(err))
			return result, errs.ErrInternalError
		}
	} else {
		activeB.Status = quiz_buzz.StatusWrong
		if err := s.buzzRepository.Save(ctx, s.db, *activeB); err != nil {
			log.Error("failed save buzz", logger.Err(err))
			return result, errs.ErrInternalError
		}

		// the fetched slice is stale — it still has this buzz as pending
		used := attemptsUsed(refreshBuzz(buzzes, *activeB))
		maxAttempts := len(active)

		s.mu.Lock()
		t, hasTimer := s.timers[eventID]
		expired := !hasTimer || t.remainingNow() <= 0
		if used >= maxAttempts || expired {
			// all competing teams failed (or time ran out) — question closes;
			// it stays used and the host moves to the next reserve question
			log.Info("tiebreak question closed",
				slog.String("event_id", eventID.String()),
				slog.Int("question", session.TiebreakQuestion),
				slog.Int("attempts_used", used))
			if hasTimer {
				t.remaining = 0
				t.running = false
			}
		} else if t.running {
			// the timer was resumed mid-answer (manual resume) — let it keep running
		} else {
			// answer failed — resume the timer; the next press opens a fresh window
			t.deadline = time.Now().Add(t.remaining)
			t.running = true
		}
		s.mu.Unlock()
		s.clearBuzzWindow(eventID)
	}

	if err := s.broadcastState(ctx, eventID); err != nil {
		return result, err
	}
	return result, nil
}

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
		RulesView:        normalizeRulesView(session.RulesView),
		TiebreakActive:   session.TiebreakActive,
		TiebreakQuestion: session.TiebreakQuestion,
		AttemptsMax:      MaxAttemptsRound1,
		Rounds:           RoundConfigs,
		BuzzQueue:        []quiz_dto.BuzzEntry{},
		Teams:            make([]quiz_dto.TeamState, 0, len(teams)),
	}

	// the tied group still competing: buzz whitelist, attempts cap, team flags
	inTiebreak := map[uuid.UUID]bool{}
	if session.TiebreakActive {
		rows, err := s.tiebreakRepository.GetByEvent(ctx, s.db, eventID)
		if err != nil {
			log.Error("failed get tiebreak group", logger.Err(err))
			return quiz_dto.PublicState{}, errs.ErrInternalError
		}
		active := activeTiebreakTeams(rows)
		state.AttemptsMax = len(active)
		for _, r := range active {
			inTiebreak[r.TeamID] = true
		}
	}

	cfg, cfgOk := roundConfig(session.CurrentRound)
	if session.TiebreakActive {
		t := s.getTimer(eventID, quiz_dto.RoundConfig{TimeSec: TiebreakTimeSec})
		state.Timer = quiz_dto.TimerState{
			DurationSec: TiebreakTimeSec,
			RemainingMs: t.remainingNow().Milliseconds(),
			Running:     t.running,
			Started:     t.started,
		}
	} else if cfgOk {
		t := s.getTimer(eventID, cfg)
		state.Timer = quiz_dto.TimerState{
			DurationSec: cfg.TimeSec,
			RemainingMs: t.remainingNow().Milliseconds(),
			Running:     t.running,
			Started:     t.started,
		}
	}

	if session.TiebreakActive {
		if session.TiebreakQuestion > 0 {
			question, err := s.questionRepository.GetByNumber(ctx, s.db, eventID, RoundReserve, session.TiebreakQuestion)
			if err == nil {
				qs := &quiz_dto.QuestionState{
					Number:    question.QuestionNumber,
					Text:      question.Text,
					Cancelled: question.Cancelled,
				}
				if session.AnswerRevealed {
					qs.Answer = question.Answer
					qs.Explanation = question.Explanation
					qs.Criteria = question.Criteria
				}
				state.Question = qs
			}
		}
	} else if cfgOk && session.CurrentQuestion > 0 && session.CurrentQuestion <= cfg.QuestionCount {
		question, err := s.questionRepository.GetByNumber(ctx, s.db, eventID, session.CurrentRound, session.CurrentQuestion)
		if err == nil {
			qs := &quiz_dto.QuestionState{
				Number:    question.QuestionNumber,
				Text:      question.Text,
				Cancelled: question.Cancelled,
			}
			if session.AnswerRevealed {
				qs.Answer = question.Answer
				qs.Explanation = question.Explanation
				qs.Criteria = question.Criteria
			}
			state.Question = qs
		}
	}

	buzzRound, buzzQuestion := session.CurrentRound, session.CurrentQuestion
	showQueue := session.CurrentRound == Round1 && session.CurrentQuestion > 0
	if session.TiebreakActive {
		buzzRound, buzzQuestion = RoundReserve, session.TiebreakQuestion
		showQueue = session.TiebreakQuestion > 0
	}
	if showQueue {
		buzzes, err := s.buzzRepository.GetByQuestion(ctx, s.db, eventID, buzzRound, buzzQuestion)
		if err != nil {
			log.Error("failed get buzzes", logger.Err(err))
			return quiz_dto.PublicState{}, errs.ErrInternalError
		}
		teamNames := make(map[uuid.UUID]string, len(teams))
		for _, t := range teams {
			teamNames[t.TeamID] = t.Team.Name
		}
		// show the latest window race: the fastest team plus the others with their offsets
		latestWindow := maxWindowIdx(buzzes)
		for _, b := range buzzes {
			if b.WindowIdx != latestWindow {
				continue
			}
			state.BuzzQueue = append(state.BuzzQueue, quiz_dto.BuzzEntry{
				TeamID:   b.TeamID.String(),
				TeamName: teamNames[b.TeamID],
				Status:   b.Status,
				IsActive: b.Status == quiz_buzz.StatusPending,
				OffsetMs: b.OffsetMs,
			})
		}
		state.AttemptsUsed = attemptsUsed(buzzes)
	}

	// the group shrinks as teams win steps; never show fewer attempts than used
	if session.TiebreakActive && state.AttemptsMax < state.AttemptsUsed {
		state.AttemptsMax = state.AttemptsUsed
	}

	for i, t := range teams {
		r1 := t.Round1Score + t.Round1Adjust
		r2 := t.Round2Score + t.Round2Adjust
		r3 := t.Round3Score + t.Round3Adjust
		total := r1 + r2 + r3
		livesLeft := t.TotalAllowedMistakes() - t.Round3WrongCount
		if livesLeft < 0 {
			livesLeft = 0
		}
		state.Teams = append(state.Teams, quiz_dto.TeamState{
			TeamID:           t.TeamID.String(),
			Number:           i + 1,
			Name:             t.Team.Name,
			Round1Score:      r1,
			Round2Score:      r2,
			Round3Score:      r3,
			TotalScore:       total,
			Round3Order:      t.Round3Order,
			Round3WrongCount: t.Round3WrongCount,
			Round3LivesLeft:  livesLeft,
			Round3Eliminated: t.Round3Eliminated,
			TiebreakScore:    t.TiebreakScore,
			InTiebreak:       inTiebreak[t.TeamID],
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
