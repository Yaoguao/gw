package exchanger_test

import (
	"context"
	"gw-currency-wallet/internal/services/exchanger"
	"gw-currency-wallet/internal/services/exchanger/mocks"
	pgxdriver "gw-currency-wallet/pkg/pgx-driver"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetExchangeRates_InvalidBase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	getter := mocks.NewMockGetterExchangerRates(ctrl)

	service := exchanger.NewServiceExchanger(
		nil,
		slog.Default(),
		getter,
		nil,
		nil,
		nil,
		time.Minute,
	)

	_, err := service.GetExchangeRates(context.Background(), "EU")

	assert.Error(t, err)
}

func TestGetExchangeRates_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	getter := mocks.NewMockGetterExchangerRates(ctrl)

	rates := map[string]float64{
		"USD": 1.1,
		"RUB": 90,
	}

	getter.
		EXPECT().
		GetExchangeRates(gomock.Any(), "EUR").
		Return(rates, nil)

	service := exchanger.NewServiceExchanger(
		nil,
		slog.Default(),
		getter,
		nil,
		nil,
		nil,
		time.Minute,
	)

	res, err := service.GetExchangeRates(context.Background(), "EUR")

	assert.NoError(t, err)
	assert.Equal(t, rates, res)
}

func TestGetExchangeRate_CacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	getter := mocks.NewMockGetterExchangerRates(ctrl)

	service := exchanger.NewServiceExchanger(
		nil,
		slog.Default(),
		getter,
		nil,
		nil,
		nil,
		time.Minute,
	)

	getter.EXPECT().GetExchangeRateForCurrency(gomock.Any(), "USD", "EUR").Return(int64(100), nil)
	rate1, _ := service.GetExchangeRate("USD", "EUR")
	rate2, _ := service.GetExchangeRate("USD", "EUR")

	//assert.NoError(t, err)
	assert.Equal(t, rate1, rate2)
}

func TestExchange_InvalidCurrency(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := exchanger.NewServiceExchanger(
		nil,
		slog.Default(),
		nil,
		nil,
		nil,
		nil,
		time.Minute,
	)

	_, err := service.Exchange(
		context.Background(),
		uuid.New(),
		"US",
		"EUR",
		100,
	)

	assert.Error(t, err)
}

func TestExchange_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	getter := mocks.NewMockGetterExchangerRates(ctrl)
	saverEx := mocks.NewMockSaverExchanger(ctrl)
	balance := mocks.NewMockBalanceUpdaterWallet(ctrl)
	txManager := mocks.NewMockManager(ctrl)

	userID := uuid.New()

	getter.
		EXPECT().
		GetExchangeRateForCurrency(gomock.Any(), "USD", "EUR").
		Return(int64(90), nil)

	txManager.
		EXPECT().
		ExecuteInTransaction(gomock.Any(), "exchange", gomock.Any()).
		DoAndReturn(func(ctx context.Context, name string, fn func(pgxdriver.QueryExecuter) error) error {
			return fn(nil)
		})

	balance.
		EXPECT().
		DecreaseBalance(gomock.Any(), nil, userID, "USD", int64(100)).
		Return(int64(0), nil)

	balance.
		EXPECT().
		IncreaseBalance(gomock.Any(), nil, userID, "EUR", int64(90)).
		Return(int64(90), nil)

	saverEx.
		EXPECT().
		CreateExchange(gomock.Any(), nil, gomock.Any()).
		Return(nil)

	service := exchanger.NewServiceExchanger(
		txManager,
		slog.Default(),
		getter,
		saverEx,
		balance,
		nil,
		time.Minute,
	)

	res, err := service.Exchange(
		context.Background(),
		userID,
		"USD",
		"EUR",
		100,
	)

	assert.NoError(t, err)
	assert.Equal(t, int64(90), res)
}
