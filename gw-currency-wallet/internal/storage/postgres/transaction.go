package postgres

import (
	"context"
	"gw-currency-wallet/internal/domain/models"
	pgxdriver "gw-currency-wallet/pkg/pgx-driver"
	"gw-currency-wallet/pkg/pgx-driver/transaction"
	"log/slog"
)

type TransactionRepository struct {
	postgres *pgxdriver.Postgres
	log      *slog.Logger
}

func NewTransactionRepository(log *slog.Logger, postgres *pgxdriver.Postgres) *TransactionRepository {
	return &TransactionRepository{
		postgres: postgres,
		log:      log,
	}
}

func (r *TransactionRepository) CreateTransaction(
	ctx context.Context,
	tx pgxdriver.QueryExecuter,
	operation *models.Transaction,
) error {

	const op = "storage.postgres.CreateTransaction"

	query, args, err := r.postgres.Insert("transactions").
		Columns("id", "user_id", "type", "currency", "amount").
		Values(operation.ID, operation.UserID, operation.Type, operation.Currency, operation.Amount).
		ToSql()

	if err != nil {
		return transaction.HandleError(op, "insert", err)
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return transaction.HandleError(op, "insert", err)
	}

	return nil
}
