package app

import (
	"context"
	"fmt"
	"log/slog"
	"main/internal/SSO/handlers/session_handler"
	"main/internal/SSO/handlers/user_handler"
	session_repository "main/internal/SSO/repository/session-repository"
	user_repository "main/internal/SSO/repository/user-repository"
	sso_service "main/internal/SSO/service"
	"main/internal/game/repository/game_participant_repository"
	"main/internal/game/repository/game_repository"
	"main/internal/game/service/game_service"
	"main/internal/leaderboard/repository/leaderboard_repository"

	"main/internal/config"
	"main/internal/event/handler/event_handler"
	event_repository "main/internal/event/repository"
	event_service "main/internal/event/service"
	"main/internal/participant/repository/participant_repository"
	"main/internal/pkg/logger"
	"main/internal/storage/postgresql"
	"main/internal/team/handler/team_handler"
	"main/internal/team/repository/team_participant_repository"
	"main/internal/team/repository/team_repository"
	"main/internal/team/service/team_service"
	"net/http"
	"time"

	auth_verification "main/internal/pkg/middleware/authVerification"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

type App struct {
	server *http.Server
	log    *slog.Logger
}

func New(log *slog.Logger, cfg *config.Config) *App {
	// init storage
	storage := postgresql.MustPostgreSQL(cfg.GetDSN())

	// init repository
	gameRepository := game_repository.New()
	gameParticpantRepository := game_participant_repository.New()
	teamRepository := team_repository.New()
	participantRepository := participant_repository.New()
	eventRepository := event_repository.New()
	teamParticipantRepository := team_participant_repository.New()
	leaderboardRepository := leaderboard_repository.New()
	userRepository := user_repository.New()
	sessionRepository := session_repository.New()

	// init service
	gameService := game_service.New(log, storage, gameRepository, leaderboardRepository, gameParticpantRepository, teamParticipantRepository)
	teamService := team_service.New(teamRepository, participantRepository, teamParticipantRepository, log, storage)
	eventService := event_service.New(eventRepository, log, storage, gameService)
	ssoService := sso_service.New(storage, log, userRepository, sessionRepository)

	// init handler
	teamHandler := team_handler.New(teamService)
	eventHandler := event_handler.New(eventService)
	sessionHandler := session_handler.NewSessionHandler(ssoService)
	userHandler := user_handler.NewUserHandler(ssoService)

	//init auth
	// init auth middleware
	authMiddleware := auth_verification.NewAuthMiddleware(sessionRepository, storage)

	var app = &App{
		log: log,
	}

	// init router
	router := chi.NewRouter()
	app.initRoutes(router, teamHandler, eventHandler, userHandler, sessionHandler, authMiddleware)

	app.server = &http.Server{
		Addr:    cfg.GetAddress(),
		Handler: router,
	}

	return app
}

func (a *App) initRoutes(router *chi.Mux, teamHandler *team_handler.TeamHandler, eventHandler *event_handler.EventHandler, userHandler *user_handler.UserHandler, sessionHandler *session_handler.SessionHandler, authMiddleware *auth_verification.AuthMiddleware) {

	// Группа маршрутов для аутентификации и сессий
	router.Route("/api/sessions", func(r chi.Router) {
		r.Post("/login", sessionHandler.Login)        // POST /api/sessions/login
		r.Post("/refresh", sessionHandler.Refresh)    // POST /api/sessions/refresh
		r.Post("/logout/{id}", sessionHandler.Logout) // POST /api/sessions/logout/{id}
		r.Get("/{id}", sessionHandler.GetSessionByID) // GET /api/sessions/{id}

		// Применяем middleware только к конкретному роуту
		r.With(authMiddleware.AuthVerify).Get("/", sessionHandler.ListSessions) // GET /api/sessions
		// r.Put("/{id}", sessionHandler.UpdateSession)  // PUT /api/sessions/{id}
	})

	// Группа маршрутов для пользователей
	router.Route("/api/users", func(r chi.Router) {
		r.Post("/", userHandler.CreateUser)       // POST /api/users
		r.Get("/{id}", userHandler.GetUserByID)   // GET /api/users/{id}
		r.Get("/", userHandler.ListUsers)         // GET /api/users
		r.Put("/{id}", userHandler.UpdateUser)    // PUT /api/users/{id}
		r.Delete("/{id}", userHandler.DeleteUser) // DELETE /api/users/{id}
	})

	// Группа маршрутов для команд
	router.Route("/api/teams", func(r chi.Router) {
		r.Post("/register", teamHandler.RegisterTeam)       // POST /api/teams/register
		r.Get("/event/{id}", teamHandler.GetTeamsByEventId) // GET /api/teams/event/{id}
	})

	// Группа маршрутов для событий
	router.Route("/api/events", func(r chi.Router) {
		r.Post("/", eventHandler.Create)                           // POST /api/events
		r.Get("/actual", eventHandler.GetActualEvents)             // GET /api/events/actual
		r.Get("/{id}/stages", eventHandler.GetAllStagesByParentId) // GET /api/events/{id}/stages
	})
}

func (a *App) initCors(router *chi.Mux, frontedUrl string) {
	corsOptions := cors.Options{
		AllowedOrigins: []string{frontedUrl},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Requested-With",
		},
		ExposedHeaders: []string{
			"Link",
			"Content-Length",
			"Access-Control-Allow-Origin",
			"Access-Control-Allow-Credentials",
		},
		AllowCredentials: true,
		MaxAge:           300,
	}
	router.Use(cors.Handler(corsOptions))
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "app.Run"
	a.log.Info("server starting")
	if err := a.server.ListenAndServe(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.log.Error("failee to stop sever", logger.Err(err))
		return
	}
	a.log.Info("server stopped")
}
