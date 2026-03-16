package rabbitmqapp

import (
	"gw-currency-wallet/internal/config"
	"gw-currency-wallet/pkg/rabbitmq"
	"log/slog"
	"time"
)

type App struct {
	Client *rabbitmq.RabbitClient
}

func New(log *slog.Logger, cfg *config.Config) *App {
	strategy := rabbitmq.Strategy{
		Attempts: 5,
		Delay:    2 * time.Second,
		Backoff:  2,
	}

	rabbitCfg := rabbitmq.ClientConfig{
		URL:            cfg.RabbitMQ.URL,
		ConnectionName: "gw-wallet",
		ConnectTimeout: 5 * time.Second,
		Heartbeat:      10 * time.Second,
		ReconnectStrat: strategy,
		ProducingStrat: strategy,
		ConsumingStrat: strategy,
	}

	client, err := rabbitmq.NewClient(rabbitCfg)

	if err != nil {
		log.Error("error", err.Error())
		panic(err)
	}

	return &App{
		Client: client,
	}
}

func (a *App) Close() {
	err := a.Client.Close()
	if err != nil {
		return
	}
}
