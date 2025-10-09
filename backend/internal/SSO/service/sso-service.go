package sso_service

import (
	"context"
	"fmt"
	"log/slog"
	sso_dto "main/internal/SSO/dto"
	session_storage "main/internal/SSO/model/session-storage"
	user "main/internal/SSO/model/user"
	"main/internal/pkg/errs"
	"main/internal/pkg/logger"
	utils "main/internal/pkg/util"
	"time"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, db *gorm.DB, user *user.User) error
	GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (user.User, error)
	GetByLogin(ctx context.Context, db *gorm.DB, login string) (user.User, error)
	GetAll(ctx context.Context, db *gorm.DB) ([]user.User, error)
	Update(ctx context.Context, db *gorm.DB, user user.User) error
	DeleteByID(ctx context.Context, db *gorm.DB, id uuid.UUID) error
}

type UserSessionRepository interface {
	Create(ctx context.Context, db *gorm.DB, session *session_storage.UserSession) error
	GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (session_storage.UserSession, error)
	GetByUserID(ctx context.Context, db *gorm.DB, userID uuid.UUID) (session_storage.UserSession, error)
	GetByToken(ctx context.Context, db *gorm.DB, token string) (session_storage.UserSession, error)
	GetAll(ctx context.Context, db *gorm.DB) ([]session_storage.UserSession, error)
	Update(ctx context.Context, db *gorm.DB, session session_storage.UserSession) error
	DeleteByID(ctx context.Context, db *gorm.DB, id uuid.UUID) error
	DeleteByToken(ctx context.Context, db *gorm.DB, token string) error
	DeleteExpiredSessions(ctx context.Context, db *gorm.DB) error
}

// сервис
type SsoService struct {
	db                *gorm.DB
	log               *slog.Logger
	userRepository    UserRepository
	sessionRepository UserSessionRepository
}

func New(db *gorm.DB, log *slog.Logger, userRepo UserRepository, sessionRepo UserSessionRepository) *SsoService {
	alog := log.With("owner", "ssoService")
	return &SsoService{
		db:                db,
		log:               alog,
		userRepository:    userRepo,
		sessionRepository: sessionRepo,
	}
}

// ----------------- Работа с пользователями -----------------

func (s *SsoService) CreateUser(ctx context.Context, dto sso_dto.UserResponceDTO) (uuid.UUID, error) {
	const op = "services.SSOService.CreateUser"
	log := s.log.With(
		slog.String("op", op),
	)

	hashedPassword, err := utils.HashPassword(dto.Password)
	if err != nil {
		log.Error("failed hash password", logger.Err(err))
		return uuid.Nil, fmt.Errorf("hash password: %w", err)
	}

	u := user.User{
		Login:    dto.Login,
		Password: hashedPassword,
	}
	return uuid.UUID{}, s.userRepository.Create(ctx, s.db, &u)
}

func (s *SsoService) GetUserByID(ctx context.Context, id string) (sso_dto.UserResponseDTO, error) {
	const op = "services.SSOService.GetUserByID"
	log := s.log.With(
		slog.String("op", op),
	)
	uid, err := uuid.Parse(id)
	if err != nil {
		log.Error("failed parse id to uuid", logger.Err(err))
		return sso_dto.UserResponseDTO{}, errs.ErrBadRequest.Wrap("failed parse uuid")
	}

	entity, err := s.userRepository.GetByID(ctx, s.db, uid)
	if err != nil {
		return sso_dto.UserResponseDTO{}, fmt.Errorf("get user by id: %w", err)
	}

	return sso_dto.UserResponseDTO{
		ID:    entity.ID,
		Login: entity.Login,
	}, nil
}

func (s *SsoService) GetUserByLogin(ctx context.Context, login string) (sso_dto.UserResponseDTO, error) {
	const op = "services.SSOService.GetUserByLogin"
	log := s.log.With(
		slog.String("op", op),
	)
	entity, err := s.userRepository.GetByLogin(ctx, s.db, login)
	if err != nil {
		log.Error("failed get user by login", logger.Err(err))
		return sso_dto.UserResponseDTO{}, fmt.Errorf("get user by login: %w", err)
	}

	return sso_dto.UserResponseDTO{
		ID:    entity.ID,
		Login: entity.Login,
	}, nil
}

func (s *SsoService) ListUsers(ctx context.Context) ([]sso_dto.UserResponseDTO, error) {

	const op = "services.SSOService.ListUsers"
	log := s.log.With(
		slog.String("op", op),
	)

	entities, err := s.userRepository.GetAll(ctx, s.db)
	if err != nil {
		log.Error("failed get list users", logger.Err(err))
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

func (s *SsoService) UpdateUser(ctx context.Context, dto sso_dto.UpdateUserDTO) error {

	const op = "services.SSOService.UpdateUser"
	log := s.log.With(
		slog.String("op", op),
	)
	uid, err := uuid.Parse(dto.Id)
	if err != nil {
		log.Error("failed parse id to uuid", logger.Err(err))
		return errs.ErrBadRequest.Wrap("failed parse uuid")
	}

	entity, err := s.userRepository.GetByID(ctx, s.db, uid)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	if dto.Login != "" {
		entity.Login = dto.Login
	}
	if dto.Password != "" {
		hashedPassword, err := utils.HashPassword(dto.Password)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}
		entity.Password = hashedPassword
	}

	return s.userRepository.Update(ctx, s.db, entity)
}

func (s *SsoService) DeleteUser(ctx context.Context, id string) error {
	const op = "services.SSOService.DeleteUser"
	log := s.log.With(
		slog.String("op", op),
	)

	uid, err := uuid.Parse(id)
	if err != nil {
		log.Error("failed parse id to uuid", logger.Err(err))
		return errs.ErrBadRequest.Wrap("failed parse uuid")
	}

	return s.userRepository.DeleteByID(ctx, s.db, uid)
}

// ----------------- Работа с сессиями -----------------

func (s *SsoService) CreateSession(ctx context.Context, userId string) (uuid.UUID, error) {

	const op = "services.SSOService.CreateSession"
	log := s.log.With(
		slog.String("op", op),
	)

	uid, err := uuid.Parse(userId)
	if err != nil {
		log.Error("failed parse id to uuid", logger.Err(err))
		return uid, errs.ErrBadRequest.Wrap("failed parse uuid")
	}

	sessionToken := utils.GenerateToken(10)
	csrfToken := utils.GenerateToken(10)

	session := session_storage.UserSession{
		UserId:       uid,
		SessionToken: sessionToken,
		CSRFToken:    csrfToken,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(30 * time.Minute),
	}
	return uid, s.sessionRepository.Create(ctx, s.db, &session)
}

func (s *SsoService) GetSessionByID(ctx context.Context, id string) (session_storage.UserSession, error) {

	const op = "services.SSOService.GetSessionByID"
	log := s.log.With(
		slog.String("op", op),
	)

	uid, err := uuid.Parse(id)
	if err != nil {
		log.Error("failed parse id to uuid", logger.Err(err))
		return session_storage.UserSession{}, errs.ErrBadRequest.Wrap("failed parse uuid")
	}
	return s.sessionRepository.GetByID(ctx, s.db, uid)
}

func (s *SsoService) GetSessionByToken(ctx context.Context, token string) (session_storage.UserSession, error) {
	return s.sessionRepository.GetByToken(ctx, s.db, token)
}

func (s *SsoService) ListSessions(ctx context.Context) ([]session_storage.UserSession, error) {
	return s.sessionRepository.GetAll(ctx, s.db)
}

func (s *SsoService) UpdateSession(ctx context.Context, token string, dto sso_dto.UpdateUserSessionDTO) error {
	entity, err := s.sessionRepository.GetByToken(ctx, s.db, token)
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

func (s *SsoService) DeleteSession(ctx context.Context, id string) error {
	const op = "services.SSOService.DeleteSession"
	log := s.log.With(
		slog.String("op", op),
	)

	uid, err := uuid.Parse(id)
	if err != nil {
		log.Error("failed parse id to uuid", logger.Err(err))
		return errs.ErrBadRequest.Wrap("failed parse uuid")
	}
	return s.sessionRepository.DeleteByID(ctx, s.db, uid)
}

func (s *SsoService) DeleteSessionByToken(ctx context.Context, token string) error {
	return s.sessionRepository.DeleteByToken(ctx, s.db, token)
}

func (s *SsoService) DeleteExpiredSessions(ctx context.Context) error {
	return s.sessionRepository.DeleteExpiredSessions(ctx, s.db)
}

// ----------------- Аутентификация -----------------

func (s *SsoService) Login(ctx context.Context, login, password string) (uuid.UUID, string, string, error) {
	// 1. Ищем пользователя
	userEntity, err := s.userRepository.GetByLogin(ctx, s.db, login)
	if err != nil {
		return uuid.Nil, "", "", fmt.Errorf("login: user not found: %w", err)
	}

	// 2. Проверяем пароль
	if !utils.CheckPasswordHash(password, userEntity.Password) {
		return uuid.Nil, "", "", fmt.Errorf("login: invalid credentials")
	}

	// 3. Генерируем токены
	sessionToken := utils.GenerateToken(10)
	csrfToken := utils.GenerateToken(10)

	//4 Проверяем есть ли уже сессия у пользователя, если есть, удаляем её
	existingSession, err := s.sessionRepository.GetByUserID(ctx, s.db, userEntity.ID)

	if err == nil {
		// Session exists, delete it
		if err := s.sessionRepository.DeleteByID(ctx, s.db, existingSession.ID); err != nil {
			return uuid.Nil, "", "", fmt.Errorf("login: failed to delete existing session: %w", err)
		}
	} else if err != gorm.ErrRecordNotFound {
		// Unexpected error
		return uuid.Nil, "", "", fmt.Errorf("login: failed to get existing session: %w", err)
	}

	// 5. Создаем новую сессию
	session := session_storage.UserSession{
		UserId:       userEntity.ID,
		SessionToken: sessionToken,
		CSRFToken:    csrfToken,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(30 * time.Minute),
	}

	err = s.sessionRepository.Create(ctx, s.db, &session)
	if err != nil {
		return uuid.Nil, "", "", fmt.Errorf("login: failed to create session: %w", err)
	}

	return session.ID, sessionToken, csrfToken, nil
}

func (s *SsoService) Refresh(ctx context.Context, token string) (uuid.UUID, string, string, error) {
	// 1. Проверяем существование сессии
	sessionEntity, err := s.sessionRepository.GetByToken(ctx, s.db, token)
	if err != nil {
		return uuid.UUID{}, "", "", fmt.Errorf("refresh: session not found: %w", err)
	}

	// 2. Проверяем не истекла ли сессия
	if time.Now().After(sessionEntity.ExpiresAt) {
		return uuid.UUID{}, "", "", fmt.Errorf("refresh: session expired")
	}

	// 3. Генерируем новые токены
	sessionToken := utils.GenerateToken(10) // рекомендуется более длинный токен
	csrfToken := utils.GenerateToken(10)

	// 4. Обновляем сущность
	sessionEntity.SessionToken = sessionToken
	sessionEntity.CSRFToken = csrfToken
	sessionEntity.UpdatedAt = time.Now()
	sessionEntity.ExpiresAt = time.Now().Add(30 * time.Minute)

	// 5. Сохраняем обновленную сессию
	if err := s.sessionRepository.Update(ctx, s.db, sessionEntity); err != nil {
		return uuid.UUID{}, "", "", fmt.Errorf("refresh: failed to update session: %w", err)
	}

	return sessionEntity.UserId, sessionToken, csrfToken, nil
}

func (s *SsoService) Logout(ctx context.Context, sessionToken string) error {
	// Удаляем сессию по ID
	if err := s.sessionRepository.DeleteByToken(ctx, s.db, sessionToken); err != nil {
		return fmt.Errorf("logout: failed to delete session: %w", err)
	}
	return nil
}
