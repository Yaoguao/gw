package mongodb

import (
	"context"
	"fmt"
	"gw-notification/internal/config"
	"gw-notification/internal/domain/models"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoStorage struct {
	log *slog.Logger

	client *mongo.Client
	col    *mongo.Collection
}

func NewMongoStorage(cfg *config.Config, log *slog.Logger) (*MongoStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(cfg.Storage.MongoDB.DSN)
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("ping mongo: %w", err)
	}

	db := client.Database(cfg.Storage.MongoDB.DBName)
	col := db.Collection("translation")

	return &MongoStorage{
		log:    log,
		client: client,
		col:    col,
	}, nil
}

func (m *MongoStorage) Close() {
	err := m.client.Disconnect(context.Background())
	if err != nil {
		return
	}
}

func (m *MongoStorage) CreateTranslation(ctx context.Context, translation models.LargeTranslation) error {

	m.log.Debug("CreateTranslation model", translation)

	_, err := m.col.InsertOne(ctx, translation)

	if err != nil {
		return fmt.Errorf("mongo: insert translation: %w", err)
	}

	return nil
}
