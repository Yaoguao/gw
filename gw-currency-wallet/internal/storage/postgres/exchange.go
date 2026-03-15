package postgres

import (
	"context"
	"gw-currency-wallet/internal/domain/models"
	pgxdriver "gw-currency-wallet/pkg/pgx-driver"
	"gw-currency-wallet/pkg/pgx-driver/transaction"
	"log/slog"
)

type ExchangeRepository struct {
	postgres *pgxdriver.Postgres
	log      *slog.Logger
}

func NewExchangeRepository(log *slog.Logger, postgres *pgxdriver.Postgres) *ExchangeRepository {
	return &ExchangeRepository{
		postgres: postgres,
		log:      log,
	}
}

func (r *ExchangeRepository) CreateExchange(
	ctx context.Context,
	tx pgxdriver.QueryExecuter,
	exchange *models.Exchange,
) error {

	const op = "storage.postgres.CreateExchange"

	query, args, err := r.postgres.Insert("exchanges").
		Columns("id", "user_id", "from_currency", "to_currency", "amount_from", "amount_to", "rate").
		Values(
			exchange.ID,
			exchange.UserID,
			exchange.FromCurrency,
			exchange.ToCurrency,
			exchange.AmountFrom,
			exchange.AmountTo,
			exchange.Rate).
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
