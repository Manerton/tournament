package sso_dto

import (
	"main/internal/SSO/model/user"

	"github.com/google/uuid"
)

// DTO для создания пользователя
type UserResponceDTO struct {
	Login    string        `json:"login" validate:"required"`
	Password string        `json:"password" validate:"required"`
	Role     user.RoleType `json:"role,omitempty"`
}

// DTO для обновления пользователя
type UpdateUserDTO struct {
	Id       string        `json:"-"`
	Login    string        `json:"login,omitempty"`
	Password string        `json:"password,omitempty"`
	Role     user.RoleType `json:"role,omitempty"`
}

// DTO для ответа
type UserResponseDTO struct {
	ID    uuid.UUID     `json:"id"`
	Login string        `json:"login"`
	Role  user.RoleType `json:"role"`
}

// DTO для обновления сессии
type UpdateUserSessionDTO struct {
	SessionToken string `json:"session_token,omitempty"`
	CSRFToken    string `json:"csrf_token,omitempty"`
}

type SessionResponseDTO struct {
	UserId       uuid.UUID     `json:"user_id"`
	Role         user.RoleType `json:"role"`
	SessionToken string        `json:"session_token,omitempty"`
	CSRFToken    string        `json:"csrf_token,omitempty"`
}
