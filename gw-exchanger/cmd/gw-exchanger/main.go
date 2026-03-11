package main

import (
	"gw-exchanger/internal/app"
	"gw-exchanger/internal/config"
	"gw-exchanger/internal/lib/logger/sl"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustLoad()

	log := sl.InitLogger(cfg.Env, os.Stdout)

	log.Debug("CONFIG", cfg)

	application := app.New(cfg, log)

	go application.GRPCServer.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.GRPCServer.Stop()

	application.Postgres.Close()
}
