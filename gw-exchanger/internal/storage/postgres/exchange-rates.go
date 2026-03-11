package postgres

import (
	"context"
	"fmt"
	"gw-exchanger/internal/domain/models"
	pgxdriver "gw-exchanger/pkg/pgx-driver"
	"gw-exchanger/pkg/pgx-driver/transaction"
	"log/slog"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type ExchangeRatesRepository struct {
	postgres *pgxdriver.Postgres
	log      *slog.Logger
}

func NewExchangeRatesRepository(log *slog.Logger, postgres *pgxdriver.Postgres) *ExchangeRatesRepository {
	return &ExchangeRatesRepository{
		postgres: postgres,
		log:      log,
	}
}

func (ex *ExchangeRatesRepository) GetExchangeRates(ctx context.Context, base string) (*models.ExchangeRatesResponse, error) {
	const op = "storage.postgres.GetExchangeRates"

	query, args, err := ex.postgres.
		Select("to_currency", "rate").
		From("exchange_rates").
		Where("from_currency = ?", base).
		ToSql()

	if err != nil {
		return nil, transaction.HandleError(op, "select", err)
	}

	rows, err := ex.postgres.Pool.Query(ctx, query, args...)
	if err != nil {
		ex.log.Error("failed to get exchange rates", "err", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	rates := make(map[string]float64)
	for rows.Next() {
		var to string
		var rate float64

		if err := rows.Scan(&to, &rate); err != nil {
			ex.log.Error("failed to scan row", "err", err)
			continue
		}
		rates[to] = rate
	}

	return &models.ExchangeRatesResponse{
		BaseCurrency: base,
		Rates:        rates,
	}, nil
}

func (ex *ExchangeRatesRepository) GetExchangeRateForCurrency(ctx context.Context, from, to string) (*models.ExchangeRates, error) {
	const op = "storage.postgres.GetExchangeRateForCurrency"

	query, args, err := ex.postgres.
		Select("rate", "updated_at").
		From("exchange_rates").
		Where(squirrel.And{
			squirrel.Expr("from_currency = ?", from),
			squirrel.Expr("to_currency = ?", to),
		}).
		ToSql()

	if err != nil {
		return nil, transaction.HandleError(op, "select", err)
	}

	var (
		rate      float64
		updatedAt time.Time
	)

	err = ex.postgres.Pool.QueryRow(ctx, query, args...).Scan(&rate, &updatedAt)
	if err != nil {
		ex.log.Error("failed to get exchange rate", "from", from, "to", to, "err", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &models.ExchangeRates{
		FromCurrency: from,
		ToCurrency:   to,
		Rate:         rate,
		UpdatedAt:    updatedAt,
	}, nil
}

func (ex *ExchangeRatesRepository) SaveBatch(ctx context.Context, pairs map[string]map[string]float64) error {

	rows := make([][]any, 0)

	for from, m := range pairs {
		for to, rate := range m {

			rows = append(rows, []any{
				from,
				to,
				rate,
				time.Now(),
			})

		}
	}

	_, err := ex.postgres.Pool.CopyFrom(
		ctx,
		pgx.Identifier{"exchange_rates"},
		[]string{"from_currency", "to_currency", "rate", "updated_at"},
		pgx.CopyFromRows(rows),
	)

	return err
}
