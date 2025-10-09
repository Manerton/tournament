package auth_verification

import (
	"errors"
	"net/http"
	"time"

	session_repository "main/internal/SSO/repository/session-repository"

	"gorm.io/gorm"
)

var AuthError = errors.New("authentication error")

type AuthMiddleware struct {
	sessionRepo *session_repository.SessionRepository
	db          *gorm.DB
}

func NewAuthMiddleware(sessionRepo *session_repository.SessionRepository, db *gorm.DB) *AuthMiddleware {
	return &AuthMiddleware{
		sessionRepo: sessionRepo,
		db:          db,
	}
}

func (m *AuthMiddleware) AuthVerify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем session_token из куки
		st, err := r.Cookie("session_token")
		if err != nil || st == nil || st.Value == "" || st.Value == "deleted" {
			http.Error(w, "Unauthorized: missing or invalid session token", http.StatusUnauthorized)
			return
		}

		// Получаем CSRF-токен из заголовка
		csrfToken := r.Header.Get("X-CSRF-Token")
		if csrfToken == "" {
			http.Error(w, "Forbidden: missing CSRF token", http.StatusForbidden)
			return
		}

		// Проверяем сессию в базе данных
		ctx := r.Context()
		session, err := m.sessionRepo.GetByToken(ctx, m.db, st.Value)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Если сессия не найдена
		if session.SessionToken == "" {
			http.Error(w, "Unauthorized: session not found", http.StatusUnauthorized)
			return
		}

		// Сравниваем CSRF-токены
		if session.CSRFToken != csrfToken {
			http.Error(w, "Forbidden: invalid CSRF token", http.StatusForbidden)
			return
		}

		// Здесь можно добавить дополнительные проверки:
		// - срок действия сессии
		// - активность пользователя
		// - и т.д.

		next.ServeHTTP(w, r)
	})
}

// Альтернативная версия без структуры
func AuthVerifyWithRepo(sessionRepo *session_repository.SessionRepository, db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			st, err := r.Cookie("session_token")
			if err != nil || st == nil || st.Value == "" || st.Value == "deleted" {
				http.Error(w, "Unauthorized: missing or invalid session token", http.StatusUnauthorized)
				return
			}

			csrfToken := r.Header.Get("X-CSRF-Token")
			if csrfToken == "" {
				http.Error(w, "Forbidden: missing CSRF token", http.StatusForbidden)
				return
			}

			ctx := r.Context()
			session, err := sessionRepo.GetByToken(ctx, db, st.Value)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if session.SessionToken == "" {
				http.Error(w, "Unauthorized: session not found", http.StatusUnauthorized)
				return
			}

			if session.CSRFToken != csrfToken {
				http.Error(w, "Forbidden: invalid CSRF token", http.StatusForbidden)
				return
			}

			// Проверка срока действия сессии
			if session.ExpiresAt.Before(time.Now()) {
				http.Error(w, "Unauthorized: session expired", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Версия с дополнительной проверкой срока действия
func (m *AuthMiddleware) AuthVerifyWithExpiry(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		st, err := r.Cookie("session_token")
		if err != nil || st == nil || st.Value == "" || st.Value == "deleted" {
			http.Error(w, "Unauthorized: missing or invalid session token", http.StatusUnauthorized)
			return
		}

		csrfToken := r.Header.Get("X-CSRF-Token")
		if csrfToken == "" {
			http.Error(w, "Forbidden: missing CSRF token", http.StatusForbidden)
			return
		}

		ctx := r.Context()
		session, err := m.sessionRepo.GetByToken(ctx, m.db, st.Value)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if session.SessionToken == "" {
			http.Error(w, "Unauthorized: session not found", http.StatusUnauthorized)
			return
		}

		if session.CSRFToken != csrfToken {
			http.Error(w, "Forbidden: invalid CSRF token", http.StatusForbidden)
			return
		}

		// Проверка срока действия сессии
		if session.ExpiresAt.Before(time.Now()) {
			http.Error(w, "Unauthorized: session expired", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
