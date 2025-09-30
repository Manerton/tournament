package sso_service

import (
	"context"
	"fmt"

	sso_dto "main/internal/SSO/dto"
	user_session "main/internal/SSO/model/session-storage"
	user "main/internal/SSO/model/user"

	"github.com/google/uuid"
	
	"gorm.io/gorm"
)





type UserRepository interface {
	Create(ctx context.Context, db *gorm.DB, user user.User) (uuid.UUID, error)
	GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (user.User, error)
	GetAll(ctx context.Context, db *gorm.DB) ([]user.User, error)
	Update(ctx context.Context, db *gorm.DB, user user.User) error
	Delete(ctx context.Context, db *gorm.DB, id uuid.UUID) error
}

type UserSessionRepository interface {
	Create(ctx context.Context, db *gorm.DB, session user_session.UserSession) (uuid.UUID, error)
	GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (user_session.UserSession, error)
	GetAll(ctx context.Context, db *gorm.DB) ([]user_session.UserSession, error)
	Update(ctx context.Context, db *gorm.DB, session user_session.UserSession) error
	Delete(ctx context.Context, db *gorm.DB, id uuid.UUID) error
}

// сервис
type SsoService struct {
	db                *gorm.DB
	userRepository    UserRepository
	sessionRepository UserSessionRepository
}

func NewSsoService(db *gorm.DB, userRepo UserRepository, sessionRepo UserSessionRepository) *SsoService {
	return &SsoService{
		db:                db,
		userRepository:    userRepo,
		sessionRepository: sessionRepo,
	}
}

// ----------------- Работа с пользователями -----------------

func (s *SsoService) CreateUser(ctx context.Context, dto sso_dto.CreateUserDTO) (uuid.UUID, error) {
	u := user.User{
		Login:    dto.Login,
		Password: dto.Password, // ⚠️ тут нужно будет хэшировать пароль!
	}
	return s.userRepository.Create(ctx, s.db, u)
}

func (s *SsoService) GetUserByID(ctx context.Context, id uuid.UUID) (sso_dto.UserResponseDTO, error) {
	entity, err := s.userRepository.GetByID(ctx, s.db, id)
	if err != nil {
		return sso_dto.UserResponseDTO{}, fmt.Errorf("get user by id: %w", err)
	}

	return sso_dto.UserResponseDTO{
		ID:    entity.ID,
		Login: entity.Login,
	}, nil
}

func (s *SsoService) ListUsers(ctx context.Context) ([]sso_dto.UserResponseDTO, error) {
	entities, err := s.userRepository.GetAll(ctx, s.db)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	result := make([]sso_dto.UserResponseDTO, 0, len(entities))
	for _, e := range entities {
		result = append(result, sso_dto.UserResponseDTO{
			ID:    e.ID,
			Login: e.Login,
		})
	}
	return result, nil
}

func (s *SsoService) UpdateUser(ctx context.Context, id uuid.UUID, dto sso_dto.UpdateUserDTO) error {
	entity, err := s.userRepository.GetByID(ctx, s.db, id)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	if dto.Login != "" {
		entity.Login = dto.Login
	}
	if dto.Password != "" {
		entity.Password = dto.Password // ⚠️ тоже хэшировать!
	}

	return s.userRepository.Update(ctx, s.db, entity)
}

func (s *SsoService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.userRepository.Delete(ctx, s.db, id)
}

// ----------------- Работа с сессиями -----------------

func (s *SsoService) CreateSession(ctx context.Context, userId uuid.UUID, sessionToken, csrfToken string) (uuid.UUID, error) {
	session := user_session.UserSession{
		UserId:       userId,
		SessionToken: sessionToken,
		CSRFToken:    csrfToken,
	}
	return s.sessionRepository.Create(ctx, s.db, session)
}

func (s *SsoService) GetSessionByID(ctx context.Context, id uuid.UUID) (user_session.UserSession, error) {
	return s.sessionRepository.GetByID(ctx, s.db, id)
}

func (s *SsoService) ListSessions(ctx context.Context) ([]user_session.UserSession, error) {
	return s.sessionRepository.GetAll(ctx, s.db)
}

func (s *SsoService) UpdateSession(ctx context.Context, id uuid.UUID, dto sso_dto.UpdateUserSessionDTO) error {
	entity, err := s.sessionRepository.GetByID(ctx, s.db, id)
	if err != nil {
		return fmt.Errorf("update session: %w", err)
	}

	if dto.SessionToken != "" {
		entity.SessionToken = dto.SessionToken
	}
	if dto.CSRFToken != "" {
		entity.CSRFToken = dto.CSRFToken
	}

	return s.sessionRepository.Update(ctx, s.db, entity)
}

func (s *SsoService) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return s.sessionRepository.Delete(ctx, s.db, id)
}
