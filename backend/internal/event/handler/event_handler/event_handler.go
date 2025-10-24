package event_handler

import (
	"context"
	"main/internal/event/dto/event_dto"
	"main/internal/pkg/errs"
	"main/internal/pkg/response"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type EventServiceInterface interface {
	GetByID(context.Context, string) (event_dto.EventDTOResponse, error)
	GetAll(context.Context) ([]event_dto.EventDTOResponse, error)
	GetTournamentsByStatus(context.Context, string) ([]event_dto.EventDTOResponse, error)
	GetByPreviousEventID(context.Context, string) ([]event_dto.EventDTOResponse, error)

	GetAllTournaments(context.Context) ([]event_dto.EventDTOResponse, error)

	StartTournament(ctx context.Context, eventId string) error
	FinishEvent(ctx context.Context, eventId string) error

	Create(context.Context, event_dto.EventDTOCreate) (uuid.UUID, error)
	Update(context.Context, string, event_dto.EventDTOUpdate) error
	Delete(context.Context, string) error
}

type EventHandler struct {
	eventService EventServiceInterface
}

func New(eventService EventServiceInterface) *EventHandler {
	return &EventHandler{
		eventService: eventService,
	}
}

func (h *EventHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	event, err := h.eventService.GetByID(ctx, id)
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
		Data:       event,
	})
}

func (h *EventHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	events, err := h.eventService.GetAll(ctx)
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
		Data:       events,
	})
}

func (h *EventHandler) GetTournamentsByStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	status := chi.URLParam(r, "status")

	events, err := h.eventService.GetTournamentsByStatus(ctx, status)
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
		Data:       events,
	})
}

func (h *EventHandler) GetAllTournaments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	events, err := h.eventService.GetAllTournaments(ctx)
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
		Data:       events,
	})
}

func (h *EventHandler) GetAllStagesByParentId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")

	events, err := h.eventService.GetByPreviousEventID(ctx, id)
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
		Data:       events,
	})
}

func (h *EventHandler) StartTournament(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	err := h.eventService.StartTournament(ctx, id)
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

	render.JSON(w, r, response.SuccessResponse("success started"))
}

func (h *EventHandler) FinishEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	err := h.eventService.FinishEvent(ctx, id)
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

	render.JSON(w, r, response.SuccessResponse("success finished"))
}

func (h *EventHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var dto event_dto.EventDTOCreate
	err := render.DecodeJSON(r.Body, &dto)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest.Wrap("failed decode dto")))
		return
	}

	id, err := h.eventService.Create(ctx, dto)
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
		StatusCode: http.StatusOK,
		Status:     response.StatusOK,
		Data:       id,
	})
}

func (h *EventHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")

	var dto event_dto.EventDTOUpdate
	err := render.DecodeJSON(r.Body, &dto)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest.Wrap("failed decode dto")))
		return
	}

	err = h.eventService.Update(ctx, id, dto)
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

	render.JSON(w, r, response.SuccessResponse("success update"))
}

func (h *EventHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	err := h.eventService.Delete(ctx, id)
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

	render.JSON(w, r, response.SuccessResponse("success delete"))
}
