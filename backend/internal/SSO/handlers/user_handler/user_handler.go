package user_handler

import (
	"context"
	"net/http"

	sso_dto "main/internal/SSO/dto"
	"main/internal/pkg/errs"
	"main/internal/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// Интерфейс сервиса для работы с пользователями
type UserServiceInterface interface {
	CreateUser(ctx context.Context, dto sso_dto.CreateUserDTO) (string, error)
	GetUserByID(ctx context.Context, id string) (sso_dto.UserResponseDTO, error)
	GetUserByLogin(ctx context.Context, login string) (sso_dto.UserResponseDTO, error)
	ListUsers(ctx context.Context) ([]sso_dto.UserResponseDTO, error)
	UpdateUser(ctx context.Context, id string, dto sso_dto.UpdateUserDTO) error
	DeleteUser(ctx context.Context, id string) error
}

type UserHandler struct {
	service UserServiceInterface
}

func NewUserHandler(service UserServiceInterface) *UserHandler {
	return &UserHandler{service: service}
}

// POST /users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var createDTO sso_dto.CreateUserDTO

	if err := render.DecodeJSON(r.Body, &createDTO); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest.Wrap(err.Error())))
		return
	}

	id, err := h.service.CreateUser(ctx, createDTO)
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

// GET /users/{id}
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	user, err := h.service.GetUserByID(ctx, id)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError.Wrap(err.Error())))
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status: response.StatusOK,
		Data:   user,
	})
}

// GET /users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := h.service.ListUsers(ctx)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError.Wrap(err.Error())))
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status: response.StatusOK,
		Data:   users,
	})
}

// PUT /users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	var updateDTO sso_dto.UpdateUserDTO

	if err := render.DecodeJSON(r.Body, &updateDTO); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest.Wrap(err.Error())))
		return
	}

	if err := h.service.UpdateUser(ctx, id, updateDTO); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError.Wrap(err.Error())))
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status: response.StatusOK,
		Data:   map[string]string{"message": "user updated successfully"},
	})
}

// DELETE /users/{id}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteUser(ctx, id); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError.Wrap(err.Error())))
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status: response.StatusOK,
		Data:   map[string]string{"message": "user deleted successfully"},
	})
}
