package app

import (
	"context"
	"fmt"
	"log/slog"
	"main/internal/config"
	"main/internal/participant/repository/participant_repository"
	"main/internal/pkg/logger"
	"main/internal/storage/postgresql"
	"main/internal/team/handler/team_handler"
	"main/internal/team/repository/team_participant_repository"
	"main/internal/team/repository/team_repository"
	"main/internal/team/service/team_service"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type App struct {
	server *http.Server
	log    *slog.Logger
}

func New(log *slog.Logger, cfg *config.Config) *App {
	// init storage
	storage := postgresql.MustPostgreSQL("")

	// init repository
	teamRepository := team_repository.New()
	participantRepository := participant_repository.New()
	teamParticipantRepository := team_participant_repository.New()

	// init service
	teamService := team_service.New(teamRepository, participantRepository, teamParticipantRepository, log, storage)

	// init handler
	teamHandler := team_handler.New(teamService)

	var app = &App{}

	// init router
	router := chi.NewRouter()
	app.initRoutes(router, teamHandler)

	app.server = &http.Server{
		Addr:    cfg.GetAddress(),
		Handler: router,
	}

	return app
}

func (a *App) initRoutes(router *chi.Mux, teamHandler *team_handler.TeamHandler) {
	router.Post("/backend/team/register", teamHandler.RegisterTeam)
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
