package game_handler

import (
	"context"
	"main/internal/game/dto/game_dto"
	"main/internal/pkg/errs"
	"main/internal/pkg/response"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type GameServiceInterface interface {
	FinishGame(ctx context.Context, gameID string) error
	GetByEventId(ctx context.Context, eventId string) ([]game_dto.GameDTOResponse, error)
	GetGamesAllInformation(ctx context.Context, eventId string) ([]game_dto.GameAllInfoResponse, error)

	UpdateScore(ctx context.Context, gameId string, participantId string, score string) error
}

type GameHandler struct {
	gameService GameServiceInterface
}

func New(gameService GameServiceInterface) *GameHandler {
	return &GameHandler{
		gameService: gameService,
	}
}

func (h *GameHandler) GetByEventId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	results, err := h.gameService.GetByEventId(ctx, id)
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

func (h *GameHandler) FinisheTour(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	err := h.gameService.FinishGame(ctx, id)
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

	render.JSON(w, r, response.SuccessResponse("success finish"))
}

func (h *GameHandler) GetGamesAllInformation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	result, err := h.gameService.GetGamesAllInformation(ctx, id)
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
		Data:       result,
	})
}

func (h *GameHandler) UpdateScore(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	gameId := chi.URLParam(r, "id")

	participantIdStr := r.URL.Query().Get("participant_id")
	scoreStr := r.URL.Query().Get("score")

	err := h.gameService.UpdateScore(ctx, gameId, participantIdStr, scoreStr)
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

	render.JSON(w, r, response.SuccessResponse("success update score"))

}
