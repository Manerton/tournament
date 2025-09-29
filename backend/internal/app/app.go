package app

import (
	"main/internal/location/repository/country_repository"
	"main/internal/location/service/country_service"
	"main/internal/storage/postgresql"
	"net/http"
)

type App struct {
	server *http.Server
}

func New() *App {
	// init storage
	storage := postgresql.MustPostgreSQL("")

	// init repository
	repository := country_repository.New()

	// init service
	coutnryService := country_service.New(storage, repository)
	_ = coutnryService

	var app = App{}
	app.server = &http.Server{}

	return &app
}

func (a *App) Run() error {
	return nil
}

func (a *App) Stop() {

}

func (a *App) init() {

}
