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

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type WalletRepository struct {
	postgres *pgxdriver.Postgres
	log      *slog.Logger
}

func NewWalletRepository(log *slog.Logger, postgres *pgxdriver.Postgres) *WalletRepository {
	return &WalletRepository{
		postgres: postgres,
		log:      log,
	}
}

func (r *WalletRepository) CreateWallet(ctx context.Context, walletID, userID uuid.UUID, currency string) error {
	const op = "storage.postgres.CreateWallet"

	query, args, err := r.postgres.
		Insert("wallets").
		Columns("id", "user_id", "currency").
		Values(walletID, userID, currency).
		ToSql()

	if err != nil {
		return transaction.HandleError(op, "insert", err)
	}

	_, err = r.postgres.Pool.Query(ctx, query, args...)

	if err != nil {
		r.log.Debug("Error log", err.Error())
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *WalletRepository) GetWallet(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
	const op = "storage.postgres.GetWallet"

	query, args, err := r.postgres.
		Select("id", "user_id", "currency", "balance").
		From("wallets").
		Where("id = ?", id).
		ToSql()

	if err != nil {
		return nil, transaction.HandleError(op, "select", err)
	}

	wallet := &models.Wallet{}
	err = r.postgres.Pool.QueryRow(ctx, query, args...).Scan(
		&wallet.ID, &wallet.UserID, &wallet.Currency, &wallet.Balance,
	)
	if err != nil {
		r.log.Debug(err.Error())
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrWalletNotFound
		}
		return nil, transaction.HandleError(op, "insert", err)
	}

	return wallet, nil
}

func (r *WalletRepository) GetWalletsByUser(ctx context.Context, userID uuid.UUID) ([]models.Wallet, error) {
	const op = "storage.postgres.GetWalletsUser"

	query, args, err := r.postgres.
		Select("id", "user_id", "currency", "balance").
		From("wallets").
		Where("user_id = ?", userID).
		ToSql()

	if err != nil {
		return nil, transaction.HandleError(op, "select", err)
	}

	wallets := make([]models.Wallet, 0, 5)

	rows, err := r.postgres.Pool.Query(ctx, query, args...)

	if err != nil {
		r.log.Debug("Error log", err.Error())
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	for rows.Next() {
		var w models.Wallet

		err := rows.Scan(
			&w.ID,
			&w.UserID,
			&w.Currency,
			&w.Balance,
		)

		if err != nil {
			return nil, fmt.Errorf("%s: scan error: %w", op, err)
		}

		wallets = append(wallets, w)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", op, err)
	}

	return wallets, nil
}

//func (r *WalletRepository) GetWalletsByUserCurrency(ctx context.Context, userID uuid.UUID, currency string) ([]models.Wallet, error) {
//	const op = "storage.postgres.GetWalletsByUserCurrency"
//
//	query, args, err := r.postgres.
//		Select("id", "user_id", "currency", "balance").
//		From("wallets").
//		Where(squirrel.And{
//			squirrel.Expr("user_id = ?", userID),
//			squirrel.Expr("currency = ?", currency),
//		}).
//		ToSql()
//
//	if err != nil {
//		return nil, transaction.HandleError(op, "select", err)
//	}
//
//	wallets := make([]models.Wallet, 0, 5)
//
//	rows, err := r.postgres.Pool.Query(ctx, query, args...)
//
//	if err != nil {
//		r.log.Debug("Error log", err.Error())
//		return nil, fmt.Errorf("%s: %w", op, err)
//	}
//
//	for rows.Next() {
//		var w models.Wallet
//
//		err := rows.Scan(
//			&w.ID,
//			&w.UserID,
//			&w.Currency,
//			&w.Balance,
//		)
//
//		if err != nil {
//			return nil, fmt.Errorf("%s: scan error: %w", op, err)
//		}
//
//		wallets = append(wallets, w)
//	}
//
//	if err = rows.Err(); err != nil {
//		return nil, fmt.Errorf("%s: rows error: %w", op, err)
//	}
//
//	return wallets, nil
//}

func (r *WalletRepository) IncreaseBalance(
	ctx context.Context,
	tx pgxdriver.QueryExecuter,
	userID uuid.UUID,
	currency string,
	amount int64,
) (int64, error) {

	const op = "storage.postgres.IncreaseBalance"

	query, args, err := r.postgres.
		Update("wallets").
		Set("balance", squirrel.Expr("balance + ?", amount)).
		Where(squirrel.And{
			squirrel.Expr("user_id = ?", userID),
			squirrel.Expr("currency = ?", currency),
		}).
		Suffix("RETURNING balance").
		ToSql()

	if err != nil {
		return 0, transaction.HandleError(op, "build_update", err)
	}

	var newBalance int64

	err = tx.QueryRow(ctx, query, args...).Scan(&newBalance)

	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			r.log.Error("wallet not found")
			return 0, storage.ErrWalletNotFound
		}

		return 0, transaction.HandleError(op, "update", err)
	}

	return newBalance, nil
}

func (r *WalletRepository) DecreaseBalance(
	ctx context.Context,
	tx pgxdriver.QueryExecuter,
	userID uuid.UUID,
	currency string,
	amount int64,
) (int64, error) {

	const op = "storage.postgres.DecreaseBalance"

	query, args, err := r.postgres.
		Update("wallets").
		Set("balance", squirrel.Expr("balance - ?", amount)).
		Where(squirrel.And{
			squirrel.Expr("user_id = ?", userID),
			squirrel.Expr("currency = ?", currency),
			squirrel.Expr("balance >= ?", amount),
		}).
		Suffix("RETURNING balance").
		ToSql()

	if err != nil {
		return 0, transaction.HandleError(op, "build_update", err)
	}

	var newBalance int64

	err = tx.QueryRow(ctx, query, args...).Scan(&newBalance)

	if err != nil {

		r.log.Debug(op, err.Error())

		if errors.Is(err, pgx.ErrNoRows) {
			return 0, storage.ErrWalletNotFound
		}

		return 0, transaction.HandleError(op, "update", err)
	}

	return newBalance, nil
}
