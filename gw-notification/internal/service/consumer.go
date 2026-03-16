package service

import (
	"context"
	"encoding/json"
	"gw-notification/internal/domain/models"
	"log/slog"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type TranslationSaver interface {
	CreateTranslation(ctx context.Context, translation models.LargeTranslation) error
}

type ConsumerHandler struct {
	log *slog.Logger

	saver TranslationSaver
}

func New(log *slog.Logger, saver TranslationSaver) *ConsumerHandler {
	return &ConsumerHandler{
		log:   log,
		saver: saver,
	}
}

func (c *ConsumerHandler) MessageHandler(ctx context.Context, msg amqp091.Delivery) error {
	c.log.Info("handle msg", msg)

	var translation models.LargeTranslation

	err := json.Unmarshal(msg.Body, &translation)
	if err != nil {
		c.log.Error("failed to unmarshal message",
			"error", err,
		)
		return err
	}

	translation.ReceivedAt = time.Now()

	err = c.saver.CreateTranslation(ctx, translation)
	if err != nil {
		c.log.Error("failed to save translation",
			"translation_id", translation.TranslationID,
			"error", err,
		)
		return err
	}

	c.log.Info("large translation saved",
		"translation_id", translation.TranslationID,
		"amount", translation.Amount,
	)

	return nil
}
