package main

import (
	"main/internal/app"
	"main/internal/config"
	"main/internal/pkg/logger"
	"os"
	"os/signal"
	"syscall"
)

const configPath = ""

func main() {

	cfg := config.MustConfig(configPath)

	log := logger.InitLogger(cfg.EnvLevel)

	// init app
	app := app.New(log, cfg)
	// init gorutine for start server
	doneChanel := make(chan os.Signal, 1)
	signal.Notify(doneChanel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	// start app
	go app.MustRun()
	log.Info("server started")
	<-doneChanel
	// stop app
	app.Stop()
	log.Info("server stopped")
}
