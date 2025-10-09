package team_handler

import (
	"context"
	"main/internal/pkg/errs"
	"main/internal/pkg/response"
	"main/internal/team/dto/team_dto"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type TeamServiceInterface interface {
	RegisterTeam(ctx context.Context, dto team_dto.RegisterTeamRequestDTO) error
	AcceptTeam(ctx context.Context, id string) error
	GetTeamsByEventId(ctx context.Context, eventID string) ([]team_dto.TeamResponseDTO, error)
	GetAcceptTeamsByEventId(ctx context.Context, eventId string) ([]team_dto.TeamResponseDTO, error)
}

type TeamHandler struct {
	teamService TeamServiceInterface
}

func New(teamService TeamServiceInterface) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

func (h *TeamHandler) RegisterTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var registerTeamDTO team_dto.RegisterTeamRequestDTO
	err := render.DecodeJSON(r.Body, &registerTeamDTO)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest))
		return
	}

	err = h.teamService.RegisterTeam(ctx, registerTeamDTO)
	if err != nil {
		if apiErr, ok := errs.IsApiError(err); ok {
			render.Status(r, apiErr.HttpCode)
			render.JSON(w, r, response.ErrorApiResponse(apiErr))
		}

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError))
		return
	}

	render.JSON(w, r, response.SuccessResponse("success register team"))
}

func (h *TeamHandler) GetTeamsByEventId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	teams, err := h.teamService.GetTeamsByEventId(ctx, id)
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
		Status: response.StatusOK,
		Data:   teams,
	})
}

func (h *TeamHandler) GetAcceptTeamsByEventId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	teams, err := h.teamService.GetAcceptTeamsByEventId(ctx, id)
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
		Status: response.StatusOK,
		Data:   teams,
	})
}

func (h *TeamHandler) AcceptTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	err := h.teamService.AcceptTeam(ctx, id)
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

	render.JSON(w, r, response.SuccessResponse("succes accept"))
}
