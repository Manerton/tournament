package participant_handler

import (
	"context"
	"net/http"

	dto "main/internal/participant/dto"
	"main/internal/pkg/errs"
	"main/internal/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// Интерфейс сервиса, который будет использовать handler
type ParticipantServiceInterface interface {
	Create(ctx context.Context, createDTO dto.ParticipantCreateDTO) (string, error)
	GetByID(ctx context.Context, id string) (dto.ParticipantResponseDTO, error)
	GetAll(ctx context.Context) ([]dto.ParticipantResponseDTO, error)
	GetAllBySchoolID(ctx context.Context, schoolID string) ([]dto.ParticipantResponseDTO, error)
	Update(ctx context.Context, id string, updateDTO dto.UpdateParticipantDTO) error
	Delete(ctx context.Context, id string) error
}

type ParticipantHandler struct {
	service ParticipantServiceInterface
}

func NewParticipantHandler(service ParticipantServiceInterface) *ParticipantHandler {
	return &ParticipantHandler{service: service}
}

// Создание участника
func (h *ParticipantHandler) CreateParticipant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var createDTO dto.ParticipantCreateDTO

	if err := render.DecodeJSON(r.Body, &createDTO); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest.Wrap(err.Error())))
		return
	}

	id, err := h.service.Create(ctx, createDTO)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError.Wrap(err.Error())))
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status: response.StatusOK,
		Data:   map[string]string{"id": id},
	})
}

// Получение участника по ID
func (h *ParticipantHandler) GetParticipantByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	participant, err := h.service.GetByID(ctx, id)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError.Wrap(err.Error())))
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status: response.StatusOK,
		Data:   participant,
	})
}

// Получение всех участников
func (h *ParticipantHandler) GetAllParticipants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	participants, err := h.service.GetAll(ctx)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError.Wrap(err.Error())))
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status: response.StatusOK,
		Data:   participants,
	})
}

// Получение участников по школе
func (h *ParticipantHandler) GetParticipantsBySchoolID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	schoolID := chi.URLParam(r, "schoolID")

	participants, err := h.service.GetAllBySchoolID(ctx, schoolID)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError.Wrap(err.Error())))
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status: response.StatusOK,
		Data:   participants,
	})
}

// Обновление участника
func (h *ParticipantHandler) UpdateParticipant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	var updateDTO dto.UpdateParticipantDTO

	if err := render.DecodeJSON(r.Body, &updateDTO); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest.Wrap(err.Error())))
		return
	}

	if err := h.service.Update(ctx, id, updateDTO); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError.Wrap(err.Error())))
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status: response.StatusOK,
		Data:   map[string]string{"message": "participant updated successfully"},
	})
}

// Удаление участника
func (h *ParticipantHandler) DeleteParticipant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	if err := h.service.Delete(ctx, id); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError.Wrap(err.Error())))
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status: response.StatusOK,
		Data:   map[string]string{"message": "participant deleted successfully"},
	})
}
