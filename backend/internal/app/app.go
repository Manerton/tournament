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
	"main/internal/game/handler/game_handler"
	"main/internal/game/repository/game_participant_repository"
	"main/internal/game/repository/game_repository"
	"main/internal/game/service/game_service"
	leadeboard_handler "main/internal/leaderboard/handler"
	"main/internal/leaderboard/repository/leaderboard_repository"
	"main/internal/leaderboard/service/leaderboard_service"
	"main/internal/location/handler/location_handler"
	"main/internal/location/repository/district_repository"
	"main/internal/location/repository/school_repository"
	"main/internal/location/service/district_service"
	"main/internal/location/service/school_service"
	"main/internal/submission/handler/submission_handler"
	"main/internal/submission/repository/submission_repository"
	"main/internal/submission/service/submission_service"

	"main/internal/config"
	"main/internal/event/handler/event_handler"
	event_repository "main/internal/event/repository"
	event_service "main/internal/event/service"
	participant_handler "main/internal/participant/handlers"
	"main/internal/participant/repository/participant_repository"
	participant_service "main/internal/participant/service"
	"main/internal/pkg/logger"
	"main/internal/storage/postgresql"
	"main/internal/team/handler/team_handler"
	"main/internal/team/repository/team_participant_repository"
	"main/internal/team/repository/team_repository"
	"main/internal/team/service/team_service"
	"net/http"
	"time"

	auth_verification "main/internal/pkg/middleware/authVerification"
	"main/internal/pkg/middleware/midlogger"

	"github.com/go-chi/chi/middleware"
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
	schoolRepository := school_repository.New()
	distrcitRepository := district_repository.New()

	gameRepository := game_repository.New()
	gameParticpantRepository := game_participant_repository.New()
	teamRepository := team_repository.New()
	participantRepository := participant_repository.New()
	eventRepository := event_repository.New()
	teamParticipantRepository := team_participant_repository.New()
	leaderboardRepository := leaderboard_repository.New()
	userRepository := user_repository.New()
	sessionRepository := session_repository.New()
	submissionRepository := submission_repository.New()

	// init service
	schoolService := school_service.New(log, storage, schoolRepository)
	districtService := district_service.New(log, storage, distrcitRepository)

	leaderboardService := leaderboard_service.New(log, storage, leaderboardRepository, teamRepository, teamParticipantRepository, eventRepository, gameRepository, gameParticpantRepository)
	gameService := game_service.New(log, storage, gameRepository, leaderboardRepository, gameParticpantRepository, teamParticipantRepository, teamRepository, leaderboardService)
	teamService := team_service.New(teamRepository, participantRepository, teamParticipantRepository, log, storage)
	eventService := event_service.New(log, storage, eventRepository, teamRepository, gameService, leaderboardService)
	ssoService := sso_service.New(storage, log, userRepository, sessionRepository)
	participantService := participant_service.New(log, storage, participantRepository, teamParticipantRepository)
	submissionService := submission_service.New(log, storage, submissionRepository)

	// init handler
	locationHandler := location_handler.New(districtService, schoolService)
	gameHandler := game_handler.New(gameService)
	teamHandler := team_handler.New(teamService)
	eventHandler := event_handler.New(eventService)
	sessionHandler := session_handler.NewSessionHandler(ssoService)
	userHandler := user_handler.NewUserHandler(ssoService)
	participantHandler := participant_handler.NewParticipantHandler(participantService)
	leaderboardhandler := leadeboard_handler.New(leaderboardService)
	submissionHandler := submission_handler.New(submissionService)

	//init auth
	// init auth middleware
	authMiddleware := auth_verification.NewAuthMiddleware(sessionRepository, storage)

	var app = &App{
		log: log,
	}

	// init router
	router := chi.NewRouter()

	// init CORS
	app.initCors(router, cfg.FrontendURL)
	// init middleware logger
	router.Use(midlogger.New(log))
	router.Use(middleware.URLFormat)

	// init routes
	app.initRoutes(router,
		teamHandler,
		eventHandler,
		userHandler,
		sessionHandler,
		locationHandler,
		gameHandler,
		participantHandler,
		leaderboardhandler,
		submissionHandler,
		authMiddleware)

	// init server
	app.server = &http.Server{
		Addr:    cfg.GetAddress(),
		Handler: router,
	}

	return app
}

func (a *App) initRoutes(router *chi.Mux,
	teamHandler *team_handler.TeamHandler,
	eventHandler *event_handler.EventHandler,
	userHandler *user_handler.UserHandler,
	sessionHandler *session_handler.SessionHandler,
	locationHandler *location_handler.LocationHandler,
	gameHandler *game_handler.GameHandler,
	participantHandler *participant_handler.ParticipantHandler,
	leadeboardHandler *leadeboard_handler.LeaderboardHandler,
	submissionHandler *submission_handler.SubmissionHandler,

	authMiddleware *auth_verification.AuthMiddleware) {

	router.Route("/api/submissions", func(r chi.Router) {
		r.Use(authMiddleware.AuthVerify)
		r.Get("/{id}/all", submissionHandler.GetAllSubmissonsByEventId)
		r.Get("/{id}/one", submissionHandler.GetSubmissionByEventIdAndCode)
	})

	// Группа маршрутов для аутентификации и сессий
	router.Route("/api/sessions", func(r chi.Router) {
		r.Post("/login", sessionHandler.Login)            // POST /api/sessions/login
		r.Post("/refresh", sessionHandler.Refresh)        // POST /api/sessions/refresh
		r.Post("/logout/{id}", sessionHandler.Logout)     // POST /api/sessions/logout/{id}
		r.Get("/{id}", sessionHandler.GetSessionByUserID) // GET /api/sessions/{id}

		// Применяем middleware только к конкретному роуту
		r.With(authMiddleware.AuthVerify).Get("/", sessionHandler.ListSessions) // GET /api/sessions
		// r.Put("/{id}", sessionHandler.UpdateSession)  // PUT /api/sessions/{id}
	})

	// Группа маршрутов для пользователей
	router.Route("/api/users", func(r chi.Router) {
		r.With(authMiddleware.AuthVerify).Post("/", userHandler.CreateUser)       // POST/api/users/
		r.With(authMiddleware.AuthVerify).Get("/{id}", userHandler.GetUserByID)   // GET /api/users/{id}
		r.With(authMiddleware.AuthVerify).Get("/all", userHandler.ListUsers)      // GET /api/users/all
		r.With(authMiddleware.AuthVerify).Put("/{id}", userHandler.UpdateUser)    // PUT /api/users/{id}
		r.With(authMiddleware.AuthVerify).Delete("/{id}", userHandler.DeleteUser) // DELETE /api/users/{id}
	})

	// Группа маршрутов для локаций
	router.Route("/api/locations", func(r chi.Router) {
		r.Get("/districts", locationHandler.GetAllDistrict)
		r.Get("/districts/{id}/schools", locationHandler.GetSchoolsByDistrictId)
		r.Get("/school/{id}", locationHandler.GetById)
	})

	// Группа маршрутов для команд
	router.Route("/api/teams", func(r chi.Router) {
		r.Post("/register", teamHandler.RegisterTeam)                    // POST /api/teams/register
		r.Get("/event/{id}", teamHandler.GetTeamsByEventId)              // GET /api/teams/event/{id}
		r.Get("/event/{id}/accept", teamHandler.GetAcceptTeamsByEventId) // GET /api/teams/accept/event/{id}
		r.With(authMiddleware.AuthVerify).Post("/{id}/switch", teamHandler.SwitchTeamStatus)
	})

	// Группа маршрутов для игр
	router.Route("/api/games", func(r chi.Router) {
		r.Get("/{id}/tours", gameHandler.GetByEventId)
		r.Get("/{id}/all-info", gameHandler.GetGamesAllInformation)
		r.With(authMiddleware.AuthVerify).Post("/{id}/finish", gameHandler.FinisheGame)
		r.With(authMiddleware.AuthVerify).Post("/vote/{id}", gameHandler.UpdateScore)

		r.With(authMiddleware.AuthVerify).Post("/participants/by-list-id", gameHandler.GetGameParticipantsByListIDs)
	})

	// Группа маршрутов для событий
	router.Route("/api/events", func(r chi.Router) {
		r.With(authMiddleware.AuthVerify).Post("/create", eventHandler.Create) // POST /api/events
		r.With(authMiddleware.AuthVerify).Post("/{id}/start", eventHandler.StartTournament)
		r.With(authMiddleware.AuthVerify).Post("/{id}/finish", eventHandler.FinishEvent)
		r.With(authMiddleware.AuthVerify).Post("/{id}/next-tour", eventHandler.NextTour)

		r.With(authMiddleware.AuthVerify).Put("/{id}", eventHandler.Update)
		r.With(authMiddleware.AuthVerify).Delete("/{id}", eventHandler.Delete)

		r.Get("/{id}", eventHandler.GetByID)
		r.Get("/", eventHandler.GetAll)
		r.Get("/tournaments", eventHandler.GetAllTournaments)
		r.Get("/status/{status}", eventHandler.GetTournamentsByStatus) // GET /api/events/actual
		r.Get("/{id}/stages", eventHandler.GetAllStagesByParentId)     // GET /api/events/{id}/stages
	})

	router.Route("/api/participants", func(r chi.Router) {
		r.Get("/team/{id}", participantHandler.GetParticipantsByTeamId)
	})

	router.Route("/api/leaderboard", func(r chi.Router) {
		r.Get("/wins/{id}", leadeboardHandler.GetByAllWins)
		r.Get("/score/{id}", leadeboardHandler.GetByAllScore)
		r.Get("/participants/{id}", leadeboardHandler.GetAllParticipantsScore)
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
			"X-CSRF-Token",
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
