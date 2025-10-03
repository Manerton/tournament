package session_handler

import (
	"context"
	"net/http"

	sso_dto "main/internal/SSO/dto"
	"main/internal/pkg/errs"
	"main/internal/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// Интерфейс сервиса для работы с сессиями
type SessionServiceInterface interface {
	CreateSession(ctx context.Context, userId string, sessionToken, csrfToken string) (string, error)
	GetSessionByID(ctx context.Context, id string) (interface{}, error) // можно DTO
	ListSessions(ctx context.Context) ([]interface{}, error)            // можно DTO
	UpdateSession(ctx context.Context, id string, dto sso_dto.UpdateUserSessionDTO) error
	DeleteSession(ctx context.Context, id string) error
	Login(ctx context.Context, login, password string) (string, string, string, error)
	Logout(ctx context.Context, sessionID string) error
}

type SessionHandler struct {
	service SessionServiceInterface
}

func NewSessionHandler(service SessionServiceInterface) *SessionHandler {
	return &SessionHandler{service: service}
}

// POST /sessions/login
func (h *SessionHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var dto sso_dto.CreateUserDTO

	if err := render.DecodeJSON(r.Body, &dto); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest.Wrap(err.Error())))
		return
	}

	sessionID, sessionToken, csrfToken, err := h.service.Login(ctx, dto.Login, dto.Password)
	if err != nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest.Wrap(err.Error())))
		return
	}

	// Ставим session_token в HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true, // запрещает доступ из JS
		Secure:   true, // включить в проде (https)
		SameSite: http.SameSiteStrictMode,
	})

	// Возвращаем session_id + csrf_token в теле
	render.JSON(w, r, response.ApiResponse{
		Status: response.StatusOK,
		Data: map[string]string{
			"session_id": sessionID,
			"csrf_token": csrfToken,
		},
	})
}

// POST /sessions/logout/{id}
func (h *SessionHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	if err := h.service.Logout(ctx, id); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError.Wrap(err.Error())))
		return
	}

	// Стираем cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   -1, // удалить
	})

	render.JSON(w, r, response.ApiResponse{
		Status: response.StatusOK,
		Data:   map[string]string{"message": "logged out successfully"},
	})
}

// GET /sessions/{id}
func (h *SessionHandler) GetSessionByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	session, err := h.service.GetSessionByID(ctx, id)
	if err != nil {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrNotFound.Wrap(err.Error())))
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status: response.StatusOK,
		Data:   session,
	})
}

// GET /sessions
func (h *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sessions, err := h.service.ListSessions(ctx)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError.Wrap(err.Error())))
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status: response.StatusOK,
		Data:   sessions,
	})
}

// PUT /sessions/{id}
func (h *SessionHandler) UpdateSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	var updateDTO sso_dto.UpdateUserSessionDTO

	if err := render.DecodeJSON(r.Body, &updateDTO); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrBadRequest.Wrap(err.Error())))
		return
	}

	if err := h.service.UpdateSession(ctx, id, updateDTO); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError.Wrap(err.Error())))
		return
	}

	render.JSON(w, r, response.ApiResponse{
		Status: response.StatusOK,
		Data:   map[string]string{"message": "session updated successfully"},
	})
}
