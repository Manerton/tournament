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
	UpsertQuestion(ctx context.Context, eventID uuid.UUID, round int, number int, text string, answer string) error
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

	HeraldLogin(ctx context.Context, eventID uuid.UUID, code string) (quiz_dto.HeraldLoginResponse, error)
	ListTeams(ctx context.Context, eventID uuid.UUID) ([]quiz_dto.AdminTeam, error)
	Buzz(ctx context.Context, eventID uuid.UUID, code string) (quiz_dto.BuzzResult, error)
	ChooseOrder(ctx context.Context, eventID uuid.UUID, code string, order int) error
	Grade(ctx context.Context, eventID uuid.UUID, code string, grade string) (quiz_dto.GradeResult, error)

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
}

// Buzz handles buzzer presses from ESP32 buttons and manual herald buzzes.
func (h *QuizHandler) Buzz(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseEventID(w, r)
	if !ok {
		return
	}

	var req buzzRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil || req.AccessCode == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	result, err := h.quizService.Buzz(r.Context(), eventID, req.AccessCode)
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

	if err := h.quizService.UpsertQuestion(r.Context(), eventID, req.Round, req.QuestionNumber, req.Text, req.Answer); err != nil {
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
