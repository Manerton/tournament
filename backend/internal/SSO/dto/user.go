package sso_dto

import "github.com/google/uuid"

// DTO для создания пользователя
type CreateUserDTO struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// DTO для обновления пользователя
type UpdateUserDTO struct {
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
}

// DTO для ответа
type UserResponseDTO struct {
	ID    uuid.UUID `json:"id"`
	Login string    `json:"login"`
}

// DTO для обновления сессии
type UpdateUserSessionDTO struct {
	SessionToken string `json:"session_token,omitempty"`
	CSRFToken    string `json:"csrf_token,omitempty"`
}