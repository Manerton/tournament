package main

import (
	"main/internal/app"
	"main/internal/config"
	"main/internal/pkg/logger"
	"os"
	"os/signal"
	"syscall"
)

const configPath = "config-yaml/local.yaml"
const prodPath = "config-yaml/prod.yaml"
const abs = "C:/Users/Admin/goProject/tournament/backend/config-yaml/local.yaml"

func main() {
	// init config
	cfg := config.MustConfig(abs)
	// init logger
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
