package translation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gw-currency-wallet/internal/storage"
	pgxdriver "gw-currency-wallet/pkg/pgx-driver"
	"gw-currency-wallet/pkg/pgx-driver/transaction"
	"gw-currency-wallet/pkg/rabbitmq"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type Publisher interface {
	Publish(ctx context.Context, body []byte, routingKey string, opts ...rabbitmq.PublishOption) error
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

type SaverWallet interface {
	CreateWallet(ctx context.Context, walletID, userID uuid.UUID, currency string) error
}

type ServiceTranslation struct {
	log       *slog.Logger
	txManager transaction.Manager

	balanceUpdaterWallet BalanceUpdaterWallet
	saverWallet          SaverWallet
	publisher            Publisher
}

func NewServiceTranslation(
	txManager transaction.Manager,
	log *slog.Logger,
	balanceUpdaterWallet BalanceUpdaterWallet,
	saverWallet SaverWallet,
	publisher Publisher,
) *ServiceTranslation {
	return &ServiceTranslation{
		log:                  log,
		txManager:            txManager,
		balanceUpdaterWallet: balanceUpdaterWallet,
		saverWallet:          saverWallet,
		publisher:            publisher,
	}
}

func (s *ServiceTranslation) Translation(
	ctx context.Context,
	fromUser uuid.UUID,
	toUser uuid.UUID,
	amount int64,
	currency string,
) error {

	return s.txManager.ExecuteInTransaction(ctx, "translation", func(tx pgxdriver.QueryExecuter) error {

		_, err := s.balanceUpdaterWallet.DecreaseBalance(ctx, tx, fromUser, currency, amount)
		if err != nil {
			return fmt.Errorf("failed decrease balance: %w", err)
		}

		_, err = s.balanceUpdaterWallet.IncreaseBalance(ctx, tx, toUser, currency, amount)
		if err != nil {

			if errors.Is(err, storage.ErrWalletNotFound) {

				walletID := uuid.New()

				err = s.saverWallet.CreateWallet(ctx, walletID, toUser, currency)
				if err != nil {
					return fmt.Errorf("failed create wallet: %w", err)
				}

				_, err = s.balanceUpdaterWallet.IncreaseBalance(ctx, tx, toUser, currency, amount)
				if err != nil {
					return fmt.Errorf("failed increase balance: %w", err)
				}

			} else {
				return fmt.Errorf("failed increase balance: %w", err)
			}
		}

		if amount >= 30000 {

			event := map[string]interface{}{
				"translation_id": uuid.New().String(),
				"from_user_id":   fromUser.String(),
				"to_user_id":     toUser.String(),
				"amount":         amount,
				"currency":       currency,
				"created_at":     time.Now(),
			}

			body, err := json.Marshal(event)
			if err != nil {
				return err
			}

			err = s.publisher.Publish(
				ctx,
				body,
				"translation.large",
			)

			if err != nil {
				s.log.Error("failed publish large translation event", "error", err)
			}
		}

		return nil
	})
}
