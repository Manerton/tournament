package quiz_handler

import (
	"context"
	"net/http"

	"main/internal/pkg/errs"
	"main/internal/pkg/response"
	"main/internal/quiz/dto/quiz_dto"
	"main/internal/quiz/model/quiz_question"
	"main/internal/quiz/model/quiz_team"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type QuizServiceInterface interface {
	SetupTeams(ctx context.Context, eventID uuid.UUID) error
	ListQuestions(ctx context.Context, eventID uuid.UUID) ([]quiz_question.QuizQuestion, error)
	UpsertQuestion(ctx context.Context, eventID uuid.UUID, round int, number int, text string, answer string, explanation string, criteria string) error
	DeleteQuestion(ctx context.Context, eventID uuid.UUID, questionID uuid.UUID) error

	StartRound(ctx context.Context, eventID uuid.UUID, round int) error
	StartQuestion(ctx context.Context, eventID uuid.UUID) error
	PauseTimer(ctx context.Context, eventID uuid.UUID) error
	ResumeTimer(ctx context.Context, eventID uuid.UUID) error
	CloseQuestion(ctx context.Context, eventID uuid.UUID) error
	RevealAnswer(ctx context.Context, eventID uuid.UUID) error
	ShowLeaderboard(ctx context.Context, eventID uuid.UUID, show bool) error
	GrantTurn(ctx context.Context, eventID uuid.UUID, teamID uuid.UUID) error
	ResetQuestion(ctx context.Context, eventID uuid.UUID) error
	ResetTournament(ctx context.Context, eventID uuid.UUID) error
	AdjustScore(ctx context.Context, eventID uuid.UUID, teamID uuid.UUID, round int, delta int) error
	AdminSetOrder(ctx context.Context, eventID uuid.UUID, teamID uuid.UUID, order int) error

	StartTiebreak(ctx context.Context, eventID uuid.UUID, teamIDs []uuid.UUID) error
	StartTiebreakQuestion(ctx context.Context, eventID uuid.UUID) error
	EndTiebreak(ctx context.Context, eventID uuid.UUID) error
	JuryDecision(ctx context.Context, eventID uuid.UUID, teamID uuid.UUID, score int) error

	HeraldLogin(ctx context.Context, eventID uuid.UUID, code string) (quiz_dto.HeraldLoginResponse, error)
	ListTeams(ctx context.Context, eventID uuid.UUID) ([]quiz_dto.AdminTeam, error)
	Buzz(ctx context.Context, eventID uuid.UUID, code string, buttonID string) (quiz_dto.BuzzResult, error)
	ChooseOrder(ctx context.Context, eventID uuid.UUID, code string, order int) error
	Grade(ctx context.Context, eventID uuid.UUID, code string, grade string) (quiz_dto.GradeResult, error)
	ListButtons(ctx context.Context, eventID uuid.UUID) ([]quiz_dto.AdminButton, error)
	RegisterButton(ctx context.Context, eventID uuid.UUID, buttonID string, teamID uuid.UUID) error
	DeleteButton(ctx context.Context, buttonID string) error

	GetPublicState(ctx context.Context, eventID uuid.UUID) (quiz_dto.PublicState, error)
}

type QuizHandler struct {
	quizService QuizServiceInterface
}

func New(quizService QuizServiceInterface) *QuizHandler {
	return &QuizHandler{quizService: quizService}
}

func parseEventID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id := chi.URLParam(r, "id")
	uid, err := uuid.Parse(id)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return uuid.Nil, false
	}
	return uid, true
}

func renderError(w http.ResponseWriter, r *http.Request, err error) {
	if apiErr, ok := errs.IsApiError(err); ok {
		render.Status(r, apiErr.HttpCode)
		render.JSON(w, r, response.ErrorApiResponse(apiErr))
		return
	}
	render.Status(r, http.StatusInternalServerError)
	render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError))
}

// ---------- public ----------

func (h *QuizHandler) GetPublicState(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	state, err := h.quizService.GetPublicState(r.Context(), eventID)
	if err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status:     response.StatusOK,
		StatusCode: http.StatusOK,
		Data:       state,
	})
}

type buzzRequest struct {
	AccessCode string `json:"access_code"`
	ButtonID   string `json:"button_id"` // ESP32 button: preconfigured id -> team
}

// Buzz handles buzzer presses from ESP32 buttons (by button_id) and manual herald buzzes (by access_code).
func (h *QuizHandler) Buzz(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	var req buzzRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil || (req.AccessCode == "" && req.ButtonID == "") {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	result, err := h.quizService.Buzz(r.Context(), eventID, req.AccessCode, req.ButtonID)
	if err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status:     response.StatusOK,
		StatusCode: http.StatusOK,
		Data:       result,
	})
}

// ---------- herald ----------

type heraldRequest struct {
	AccessCode string `json:"access_code"`
}

func (h *QuizHandler) HeraldLogin(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	var req heraldRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil || req.AccessCode == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	result, err := h.quizService.HeraldLogin(r.Context(), eventID, req.AccessCode)
	if err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status:     response.StatusOK,
		StatusCode: http.StatusOK,
		Data:       result,
	})
}

type gradeRequest struct {
	AccessCode string `json:"access_code"`
	Grade      string `json:"grade"`
}

func (h *QuizHandler) Grade(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	var req gradeRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil || req.AccessCode == "" || req.Grade == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	result, err := h.quizService.Grade(r.Context(), eventID, req.AccessCode, req.Grade)
	if err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status:     response.StatusOK,
		StatusCode: http.StatusOK,
		Data:       result,
	})
}

type orderRequest struct {
	AccessCode string `json:"access_code"`
	Order      string `json:"order"`
}

var orderValues = map[string]int{
	"red":    quiz_team.OrderRed,
	"yellow": quiz_team.OrderYellow,
	"green":  quiz_team.OrderGreen,
}

func (h *QuizHandler) ChooseOrder(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	var req orderRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil || req.AccessCode == "" || req.Order == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	order, exists := orderValues[req.Order]
	if !exists {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	if err := h.quizService.ChooseOrder(r.Context(), eventID, req.AccessCode, order); err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.SuccessResponse("order chosen"))
}

// ---------- admin ----------

func (h *QuizHandler) SetupTeams(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	if err := h.quizService.SetupTeams(r.Context(), eventID); err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.SuccessResponse("teams setup"))
}

func (h *QuizHandler) ListTeams(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	teams, err := h.quizService.ListTeams(r.Context(), eventID)
	if err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status:     response.StatusOK,
		StatusCode: http.StatusOK,
		Data:       teams,
	})
}

func (h *QuizHandler) ListQuestions(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	questions, err := h.quizService.ListQuestions(r.Context(), eventID)
	if err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status:     response.StatusOK,
		StatusCode: http.StatusOK,
		Data:       questions,
	})
}

type upsertQuestionRequest struct {
	Round          int    `json:"round"`
	QuestionNumber int    `json:"question_number"`
	Text           string `json:"text"`
	Answer         string `json:"answer"`
	Explanation    string `json:"explanation"`
	Criteria       string `json:"criteria"`
}

func (h *QuizHandler) UpsertQuestion(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	var req upsertQuestionRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	if err := h.quizService.UpsertQuestion(r.Context(), eventID, req.Round, req.QuestionNumber, req.Text, req.Answer, req.Explanation, req.Criteria); err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.SuccessResponse("question saved"))
}

func (h *QuizHandler) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	questionID, err := uuid.Parse(chi.URLParam(r, "questionId"))
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	if err := h.quizService.DeleteQuestion(r.Context(), eventID, questionID); err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.SuccessResponse("question deleted"))
}

type startRoundRequest struct {
	Round int `json:"round"`
}

func (h *QuizHandler) StartRound(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	var req startRoundRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	if err := h.quizService.StartRound(r.Context(), eventID, req.Round); err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.SuccessResponse("round started"))
}

func (h *QuizHandler) StartQuestion(w http.ResponseWriter, r *http.Request) {
	h.simpleAdminAction(w, r, h.quizService.StartQuestion)
}

func (h *QuizHandler) PauseTimer(w http.ResponseWriter, r *http.Request) {
	h.simpleAdminAction(w, r, h.quizService.PauseTimer)
}

func (h *QuizHandler) ResumeTimer(w http.ResponseWriter, r *http.Request) {
	h.simpleAdminAction(w, r, h.quizService.ResumeTimer)
}

func (h *QuizHandler) CloseQuestion(w http.ResponseWriter, r *http.Request) {
	h.simpleAdminAction(w, r, h.quizService.CloseQuestion)
}

func (h *QuizHandler) RevealAnswer(w http.ResponseWriter, r *http.Request) {
	h.simpleAdminAction(w, r, h.quizService.RevealAnswer)
}

func (h *QuizHandler) ResetQuestion(w http.ResponseWriter, r *http.Request) {
	h.simpleAdminAction(w, r, h.quizService.ResetQuestion)
}

func (h *QuizHandler) ResetTournament(w http.ResponseWriter, r *http.Request) {
	h.simpleAdminAction(w, r, h.quizService.ResetTournament)
}

func (h *QuizHandler) simpleAdminAction(w http.ResponseWriter, r *http.Request, action func(ctx context.Context, eventID uuid.UUID) error) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	if err := action(r.Context(), eventID); err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.SuccessResponse("ok"))
}

type showLeaderboardRequest struct {
	Show bool `json:"show"`
}

func (h *QuizHandler) ShowLeaderboard(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	var req showLeaderboardRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	if err := h.quizService.ShowLeaderboard(r.Context(), eventID, req.Show); err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.SuccessResponse("leaderboard toggled"))
}

type grantRequest struct {
	TeamID string `json:"team_id"`
}

func (h *QuizHandler) GrantTurn(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	var req grantRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	teamID, err := uuid.Parse(req.TeamID)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	if err := h.quizService.GrantTurn(r.Context(), eventID, teamID); err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.SuccessResponse("turn granted"))
}

type adjustScoreRequest struct {
	TeamID string `json:"team_id"`
	Round  int    `json:"round"`
	Delta  int    `json:"delta"`
}

func (h *QuizHandler) AdjustScore(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	var req adjustScoreRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil || req.TeamID == "" || req.Round < 1 || req.Round > 3 || req.Delta == 0 {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	teamID, err := uuid.Parse(req.TeamID)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	if err := h.quizService.AdjustScore(r.Context(), eventID, teamID, req.Round, req.Delta); err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.SuccessResponse("score adjusted"))
}

// ---------- tie-break (переигровка) ----------

type startTiebreakRequest struct {
	TeamIDs []string `json:"team_ids"`
}

// StartTiebreak opens a tie-break for the selected tied group (2+ teams).
func (h *QuizHandler) StartTiebreak(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	var req startTiebreakRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil || len(req.TeamIDs) < 2 {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	teamIDs := make([]uuid.UUID, 0, len(req.TeamIDs))
	for _, raw := range req.TeamIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
			return
		}
		teamIDs = append(teamIDs, id)
	}

	if err := h.quizService.StartTiebreak(r.Context(), eventID, teamIDs); err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.SuccessResponse("tiebreak started"))
}

// StartTiebreakQuestion shows the next unused reserve question.
func (h *QuizHandler) StartTiebreakQuestion(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	if err := h.quizService.StartTiebreakQuestion(r.Context(), eventID); err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.SuccessResponse("tiebreak question started"))
}

// EndTiebreak closes the tie-break mode and clears the tied group.
func (h *QuizHandler) EndTiebreak(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	if err := h.quizService.EndTiebreak(r.Context(), eventID); err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.SuccessResponse("tiebreak ended"))
}

type juryDecisionRequest struct {
	TeamID string `json:"team_id"`
	Score  int    `json:"score"`
}

// JuryDecision manually sets a team's tiebreak score (jury fallback).
func (h *QuizHandler) JuryDecision(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	var req juryDecisionRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil || req.TeamID == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	teamID, err := uuid.Parse(req.TeamID)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	if err := h.quizService.JuryDecision(r.Context(), eventID, teamID, req.Score); err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.SuccessResponse("jury decision applied"))
}

type adminOrderRequest struct {
	TeamID string `json:"team_id"`
	Order  string `json:"order"`
}

// AdminSetOrder sets or resets a team's round-3 order (admin override of herald mistakes).
// Accepts "none" to reset the order.
func (h *QuizHandler) AdminSetOrder(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	var req adminOrderRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil || req.TeamID == "" || req.Order == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	order, exists := orderValues[req.Order]
	if !exists {
		if req.Order == "none" {
			order = quiz_team.OrderNone
		} else {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
			return
		}
	}

	teamID, err := uuid.Parse(req.TeamID)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	if err := h.quizService.AdminSetOrder(r.Context(), eventID, teamID, order); err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.SuccessResponse("order set"))
}

// ---------- buttons (ESP32) ----------

func (h *QuizHandler) ListButtons(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	buttons, err := h.quizService.ListButtons(r.Context(), eventID)
	if err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status:     response.StatusOK,
		StatusCode: http.StatusOK,
		Data:       buttons,
	})
}

type registerButtonRequest struct {
	ButtonID string `json:"button_id"`
	TeamID   string `json:"team_id"`
}

func (h *QuizHandler) RegisterButton(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	var req registerButtonRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil || req.ButtonID == "" || req.TeamID == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	teamID, err := uuid.Parse(req.TeamID)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	if err := h.quizService.RegisterButton(r.Context(), eventID, req.ButtonID, teamID); err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.SuccessResponse("button registered"))
}

func (h *QuizHandler) DeleteButton(w http.ResponseWriter, r *http.Request) {
	buttonID := chi.URLParam(r, "buttonId")
	if buttonID == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	if err := h.quizService.DeleteButton(r.Context(), buttonID); err != nil {
		renderError(w, r, err)
		return
	}

	render.JSON(w, r, response.SuccessResponse("button deleted"))
}
