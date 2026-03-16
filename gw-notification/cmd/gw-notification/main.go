package main

import (
	"context"
	"gw-notification/internal/config"
	"gw-notification/internal/lib/logger/sl"
	"gw-notification/internal/service"
	"gw-notification/internal/storage/mongodb"
	"gw-notification/pkg/rabbitmq"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.MustLoad()

	log := sl.InitLogger(cfg.Env, os.Stdout)

	log.Debug("config", cfg)

	storage, err := mongodb.NewMongoStorage(cfg)

	if err != nil {
		panic(err)
	}

	strategy := rabbitmq.Strategy{
		Attempts: 5,
		Delay:    2 * time.Second,
		Backoff:  2,
	}

	rabbitCfg := rabbitmq.ClientConfig{
		URL:            cfg.RabbitMQ.URL,
		ConnectionName: "gw-notification",
		ConnectTimeout: 5 * time.Second,
		Heartbeat:      10 * time.Second,
		ReconnectStrat: strategy,
		ProducingStrat: strategy,
		ConsumingStrat: strategy,
	}

	consumerCfg := rabbitmq.ConsumerConfig{
		Queue:       cfg.RabbitMQ.LargeTranslationQueue,
		ConsumerTag: "gw-notification",

		AutoAck: false,

		Workers:       10,
		PrefetchCount: cfg.RabbitMQ.PrefetchCount,

		Ask: rabbitmq.AskConfig{},
		Nack: rabbitmq.NackConfig{
			Requeue: true,
		},

		Args: nil,
	}

	client, err := rabbitmq.NewClient(rabbitCfg)

	if err != nil {
		panic(err)
	}

	handler := service.New(log, storage)

	consumer := rabbitmq.NewConsumer(client, consumerCfg, handler.MessageHandler)

	go func() {
		err := consumer.Start(context.Background())
		if err != nil {
			return
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	log.Info("shutting down")

	storage.Close()

	err = client.Close()
	if err != nil {
		return
	}
}
