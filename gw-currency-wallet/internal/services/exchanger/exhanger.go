package exchanger

import (
	"context"
	"errors"
	"fmt"
	"gw-currency-wallet/internal/domain/models"
	"gw-currency-wallet/internal/storage"
	pgxdriver "gw-currency-wallet/pkg/pgx-driver"
	"gw-currency-wallet/pkg/pgx-driver/transaction"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

type GetterExchangerRates interface {
	GetExchangeRateForCurrency(ctx context.Context, from, to string) (int64, error)
	GetExchangeRates(ctx context.Context, base string) (map[string]float64, error)
}

type SaverExchanger interface {
	CreateExchange(ctx context.Context, tx pgxdriver.QueryExecuter, exchange *models.Exchange) error
}

type SaverWallet interface {
	CreateWallet(ctx context.Context, walletID, userID uuid.UUID, currency string) error
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

type rateCacheItem struct {
	value     int64
	timestamp time.Time
}

type ServiceExchanger struct {
	mu       sync.RWMutex
	cache    map[string]rateCacheItem
	cacheTTL time.Duration

	txManager transaction.Manager
	log       *slog.Logger

	getterExchangerRates GetterExchangerRates
	saverExchanger       SaverExchanger
	balanceUpdaterWallet BalanceUpdaterWallet
	saverWallet          SaverWallet
}

func NewServiceExchanger(
	txManager transaction.Manager,
	log *slog.Logger,
	getterExchangerRates GetterExchangerRates,
	saverExchanger SaverExchanger,
	balanceUpdaterWallet BalanceUpdaterWallet,
	saverWallet SaverWallet,
	cacheTTL time.Duration,
) *ServiceExchanger {

	return &ServiceExchanger{
		txManager:            txManager,
		log:                  log,
		getterExchangerRates: getterExchangerRates,
		saverExchanger:       saverExchanger,
		balanceUpdaterWallet: balanceUpdaterWallet,
		saverWallet:          saverWallet,
		cache:                make(map[string]rateCacheItem),
		cacheTTL:             cacheTTL,
	}
}

func (s *ServiceExchanger) GetExchangeRates(ctx context.Context, base string) (map[string]float64, error) {

	if len(base) != 3 {
		return nil, fmt.Errorf("no correct format currency")
	}

	rates, err := s.getterExchangerRates.GetExchangeRates(ctx, base)

	if err != nil {
		return nil, fmt.Errorf("failed get exchange rates")
	}

	return rates, nil
}

func (s *ServiceExchanger) Exchange(ctx context.Context, userID uuid.UUID, from, to string, amount int64) (int64, error) {

	if len(from) != 3 || len(to) != 3 {
		return 0, fmt.Errorf("no correct format currency")
	}

	rate, err := s.GetExchangeRate(from, to)

	if err != nil {
		return 0, fmt.Errorf("failed get exchange rate for currency")
	}

	err = s.txManager.ExecuteInTransaction(ctx, "exchange", func(tx pgxdriver.QueryExecuter) error {

		_, err := s.balanceUpdaterWallet.DecreaseBalance(ctx, tx, userID, from, amount)

		if err != nil {
			return err
		}

		_, err = s.balanceUpdaterWallet.IncreaseBalance(ctx, tx, userID, to, rate)

		if errors.Is(err, storage.ErrWalletNotFound) {

			wid := uuid.New()

			if err := s.saverWallet.CreateWallet(ctx, wid, userID, to); err != nil {
				return fmt.Errorf("failed create wallet")
			}

			_, err = s.balanceUpdaterWallet.IncreaseBalance(ctx, tx, userID, to, rate)

			if err != nil {
				return fmt.Errorf("failed deposit wallet")
			}
		}

		if err != nil {
			return fmt.Errorf("failed increase balance")
		}

		ex := models.Exchange{
			ID:           uuid.New(),
			UserID:       userID,
			FromCurrency: from,
			ToCurrency:   to,
			Rate:         rate,
			AmountFrom:   amount,
			AmountTo:     rate,
		}

		return s.saverExchanger.CreateExchange(ctx, tx, &ex)
	})

	if err != nil {
		return 0, err
	}

	return rate, nil
}

func (s *ServiceExchanger) GetExchangeRate(from, to string) (int64, error) {
	key := from + "-" + to

	s.mu.RLock()
	item, ok := s.cache[key]
	s.mu.RUnlock()

	if ok {
		if time.Since(item.timestamp) < s.cacheTTL {
			s.log.Info("cache hit", "key", key)
			return item.value, nil
		}
	}

	rate, err := s.getterExchangerRates.GetExchangeRateForCurrency(context.Background(), from, to)
	if err != nil {
		return 0, err
	}

	s.mu.Lock()
	s.cache[key] = rateCacheItem{value: rate, timestamp: time.Now()}
	s.mu.Unlock()

	return rate, nil
}
