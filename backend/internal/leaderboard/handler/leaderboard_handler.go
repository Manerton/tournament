package leadeboard_handler

import (
	"context"
	leaderboard_dto "main/internal/leaderboard/dto"
	"main/internal/pkg/errs"
	"main/internal/pkg/response"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type LeaderboardServiceInterface interface {
	GetLeaderboardWins(ctx context.Context, EventID string) ([]leaderboard_dto.LeaderboardWinDTOResponse, error)
	GetLeaderboardScore(ctx context.Context, EventID string) ([]leaderboard_dto.LeaderboardDTOResponse, error)
	GetParticipantsScore(ctx context.Context, EventID string) ([]leaderboard_dto.LeaderboardPlayesDTOResponse, error)
	GetLeaderboardScoreLive(ctx context.Context, EventID string) ([]leaderboard_dto.LeaderboardDTOResponse, error)
}

type LeaderboardHandler struct {
	leaderboardService LeaderboardServiceInterface
}

func New(leaderboardService LeaderboardServiceInterface) *LeaderboardHandler {
	return &LeaderboardHandler{
		leaderboardService: leaderboardService,
	}
}

func (h *LeaderboardHandler) GetByAllWins(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	results, err := h.leaderboardService.GetLeaderboardWins(ctx, id)
	if err != nil {
		if apiErr, ok := errs.IsApiError(err); ok {
			render.Status(r, apiErr.HttpCode)
			render.JSON(w, r, response.ErrorApiResponse(apiErr))
			return
		}
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError))
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status:     response.StatusOK,
		StatusCode: http.StatusOK,
		Data:       results,
	})
}

func (h *LeaderboardHandler) GetByAllScore(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	results, err := h.leaderboardService.GetLeaderboardScoreLive(ctx, id)
	if err != nil {
		if apiErr, ok := errs.IsApiError(err); ok {
			render.Status(r, apiErr.HttpCode)
			render.JSON(w, r, response.ErrorApiResponse(apiErr))
			return
		}
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError))
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status:     response.StatusOK,
		StatusCode: http.StatusOK,
		Data:       results,
	})
}

func (h *LeaderboardHandler) GetAllParticipantsScore(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	results, err := h.leaderboardService.GetParticipantsScore(ctx, id)
	if err != nil {
		if apiErr, ok := errs.IsApiError(err); ok {
			render.Status(r, apiErr.HttpCode)
			render.JSON(w, r, response.ErrorApiResponse(apiErr))
			return
		}
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError))
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status:     response.StatusOK,
		StatusCode: http.StatusOK,
		Data:       results,
	})
}
