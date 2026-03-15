package wallet

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
)

var (
	ErrAmountNegativeValue = errors.New("amount negative value")
	ErrInvalidUserID       = errors.New("invalid argument user id")
)

type SaverWallet interface {
	CreateWallet(ctx context.Context, walletID, userID uuid.UUID, currency string) error
}

type SaverTransaction interface {
	CreateTransaction(ctx context.Context, tx pgxdriver.QueryExecuter, operation *models.Transaction) error
}

type GetterWallet interface {
	GetWalletsByUser(ctx context.Context, userID uuid.UUID) ([]models.Wallet, error)
}

type BalanceUpdaterWallet interface {
	IncreaseBalance(
		ctx context.Context,
		tx pgxdriver.QueryExecuter,
		userID uuid.UUID,
		currency string,
		amount int64,
	) (int64, error)

	DecreaseBalance(
		ctx context.Context,
		tx pgxdriver.QueryExecuter,
		userID uuid.UUID,
		currency string,
		amount int64,
	) (int64, error)
}

type ServiceWallet struct {
	txManager transaction.Manager
	log       *slog.Logger

	saverWallet          SaverWallet
	saverTransaction     SaverTransaction
	getterWallet         GetterWallet
	balanceUpdaterWallet BalanceUpdaterWallet
}

func NewServiceWallet(
	txManager transaction.Manager,
	log *slog.Logger,
	saverWallet SaverWallet,
	saverTransaction SaverTransaction,
	getterWallet GetterWallet,
	balanceUpdaterWallet BalanceUpdaterWallet,
) *ServiceWallet {

	return &ServiceWallet{
		txManager:            txManager,
		log:                  log,
		saverWallet:          saverWallet,
		saverTransaction:     saverTransaction,
		getterWallet:         getterWallet,
		balanceUpdaterWallet: balanceUpdaterWallet,
	}
}

func (s *ServiceWallet) Deposit(ctx context.Context, userID uuid.UUID, currency string, amount int64) error {

	if amount <= 0 {
		return ErrAmountNegativeValue
	}

	return s.txManager.ExecuteInTransaction(ctx, "deposit", func(tx pgxdriver.QueryExecuter) error {

		_, err := s.balanceUpdaterWallet.IncreaseBalance(ctx, tx, userID, currency, amount)

		if errors.Is(err, storage.ErrWalletNotFound) {

			wid := uuid.New()

			if err := s.saverWallet.CreateWallet(ctx, wid, userID, currency); err != nil {
				return fmt.Errorf("failed create wallet")
			}

			_, err = s.balanceUpdaterWallet.IncreaseBalance(ctx, tx, userID, currency, amount)

			if err != nil {
				return fmt.Errorf("failed deposit wallet")
			}
		}

		op := models.Transaction{
			ID:       uuid.New(),
			UserID:   userID,
			Type:     models.Deposit,
			Currency: currency,
			Amount:   amount,
		}

		return s.saverTransaction.CreateTransaction(ctx, tx, &op)
	})
}

func (s *ServiceWallet) Withdraw(ctx context.Context, userID uuid.UUID, currency string, amount int64) error {
	if amount <= 0 {
		return ErrAmountNegativeValue
	}

	return s.txManager.ExecuteInTransaction(ctx, "withdraw", func(tx pgxdriver.QueryExecuter) error {

		_, err := s.balanceUpdaterWallet.DecreaseBalance(ctx, tx, userID, currency, amount)

		if errors.Is(err, storage.ErrWalletNotFound) {

			if err != nil {
				return fmt.Errorf("not found wallet")
			}
		}

		op := models.Transaction{
			ID:       uuid.New(),
			UserID:   userID,
			Type:     models.Withdraw,
			Currency: currency,
			Amount:   amount,
		}

		return s.saverTransaction.CreateTransaction(ctx, tx, &op)
	})
}

func (s *ServiceWallet) GetWalletsBalanceByUser(ctx context.Context, userID uuid.UUID) (map[string]int64, error) {

	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	balance := make(map[string]int64)

	wallets, err := s.getterWallet.GetWalletsByUser(ctx, userID)

	if err != nil {
		s.log.Error("failed get wallets by user", err.Error())

		return nil, fmt.Errorf("failed get wallets by user")
	}

	for _, wallet := range wallets {
		balance[wallet.Currency] = wallet.Balance
	}

	return balance, nil
}
