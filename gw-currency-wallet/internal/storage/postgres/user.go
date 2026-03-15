package postgres

import (
	"context"
	"errors"
	"fmt"
	"gw-currency-wallet/internal/domain/models"
	"gw-currency-wallet/internal/storage"
	pgxdriver "gw-currency-wallet/pkg/pgx-driver"
	"gw-currency-wallet/pkg/pgx-driver/transaction"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	postgres *pgxdriver.Postgres
	log      *slog.Logger
}

func NewUserRepository(log *slog.Logger, postgres *pgxdriver.Postgres) *UserRepository {
	return &UserRepository{
		postgres: postgres,
		log:      log,
	}
}

func (r *UserRepository) SaveUser(ctx context.Context, id uuid.UUID, username, email string, passHash []byte) error {
	const op = "storage.postgres.SaveUser"

	query, args, err := r.postgres.
		Insert("users").
		Columns("id", "username", "email", "password_hash").
		Values(id, username, email, passHash).
		ToSql()

	if err != nil {
		return transaction.HandleError(op, "insert", err)
	}

	_, err = r.postgres.Pool.Exec(ctx, query, args...)

	if err != nil {
		r.log.Debug("Error log", err.Error())
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *UserRepository) GetUser(ctx context.Context, username string) (*models.User, error) {
	const op = "storage.postgres.GetUser"

	var user models.User

	query, args, err := r.postgres.
		Select("id", "email", "password_hash", "created_at").
		From("users").
		Where("username = ?", username).
		ToSql()

	err = r.postgres.Pool.QueryRow(ctx, query, args...).Scan(&user.ID, &user.Email, &user.PassHash, &user.CreatedAt)

	if err != nil {
		r.log.Debug("Error log", err.Error())

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrUserNotFound
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user.Username = username

	return &user, nil

}
