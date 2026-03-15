package main

import (
	"context"
	"gw-currency-wallet/internal/app"
	"gw-currency-wallet/internal/config"
	"gw-currency-wallet/internal/lib/logger/sl"
	"os"
	"os/signal"
	"syscall"

	_ "gw-currency-wallet/docs"
)

// @title Wallet API
// @version 1.0
// @description Currency wallet service API
// @host localhost:8081
// @BasePath /api/v1
func main() {
	cfg := config.MustLoad()

	log := sl.InitLogger(cfg.Env, os.Stdout)

	application := app.New(log, cfg)

	go application.HTTPServer.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.Postgres.Close()

	err := application.HTTPServer.Stop(context.Background())
	if err != nil {
		return
	}
}
