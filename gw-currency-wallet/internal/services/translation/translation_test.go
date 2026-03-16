package translation_test

import (
	"context"
	"encoding/json"
	"gw-currency-wallet/internal/lib/logger/sl"
	"gw-currency-wallet/internal/services/translation"
	"gw-currency-wallet/internal/services/translation/mocks"
	"gw-currency-wallet/internal/storage"
	pgxdriver "gw-currency-wallet/pkg/pgx-driver"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestTranslation_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := sl.InitLogger("dev", os.Stdout)

	fromUser := uuid.New()
	toUser := uuid.New()
	amount := int64(1000)
	currency := "USD"

	balanceUpdater := mocks.NewMockBalanceUpdaterWallet(ctrl)
	saverWallet := mocks.NewMockSaverWallet(ctrl)
	publisher := mocks.NewMockPublisher(ctrl)
	txManager := mocks.NewMockManager(ctrl)

	service := translation.NewServiceTranslation(txManager, log, balanceUpdater, saverWallet, publisher)

	txManager.EXPECT().
		ExecuteInTransaction(gomock.Any(), "translation", gomock.Any()).
		DoAndReturn(func(ctx context.Context, name string, fn func(pgxdriver.QueryExecuter) error) error {
			return fn(nil)
		})

	balanceUpdater.EXPECT().
		DecreaseBalance(gomock.Any(), nil, fromUser, currency, amount).
		Return(int64(0), nil)

	balanceUpdater.EXPECT().
		IncreaseBalance(gomock.Any(), nil, toUser, currency, amount).
		Return(int64(amount), nil)

	err := service.Translation(context.Background(), fromUser, toUser, amount, currency)
	require.NoError(t, err)
}

func TestTranslation_LargeAmount_Publish(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := sl.InitLogger("dev", os.Stdout)

	fromUser := uuid.New()
	toUser := uuid.New()
	amount := int64(50000)
	currency := "USD"

	balanceUpdater := mocks.NewMockBalanceUpdaterWallet(ctrl)
	saverWallet := mocks.NewMockSaverWallet(ctrl)
	publisher := mocks.NewMockPublisher(ctrl)
	txManager := mocks.NewMockManager(ctrl)

	service := translation.NewServiceTranslation(txManager, log, balanceUpdater, saverWallet, publisher)

	txManager.EXPECT().
		ExecuteInTransaction(gomock.Any(), "translation", gomock.Any()).
		DoAndReturn(func(ctx context.Context, name string, fn func(pgxdriver.QueryExecuter) error) error {
			return fn(nil)
		})

	balanceUpdater.EXPECT().
		DecreaseBalance(gomock.Any(), nil, fromUser, currency, amount).
		Return(int64(0), nil)

	balanceUpdater.EXPECT().
		IncreaseBalance(gomock.Any(), nil, toUser, currency, amount).
		Return(int64(amount), nil)

	publisher.EXPECT().
		Publish(gomock.Any(), gomock.Any(), "translation.large").
		DoAndReturn(func(ctx context.Context, body []byte, routingKey string, opts ...interface{}) error {
			var event map[string]interface{}
			err := json.Unmarshal(body, &event)
			require.NoError(t, err)
			require.Equal(t, fromUser.String(), event["from_user_id"])
			require.Equal(t, toUser.String(), event["to_user_id"])
			require.Equal(t, float64(amount), event["amount"].(float64))
			require.Equal(t, currency, event["currency"])
			return nil
		})

	err := service.Translation(context.Background(), fromUser, toUser, amount, currency)
	require.NoError(t, err)
}

func TestTranslation_CreateWalletIfNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := sl.InitLogger("dev", os.Stdout)

	fromUser := uuid.New()
	toUser := uuid.New()
	amount := int64(1000)
	currency := "USD"

	balanceUpdater := mocks.NewMockBalanceUpdaterWallet(ctrl)
	saverWallet := mocks.NewMockSaverWallet(ctrl)
	publisher := mocks.NewMockPublisher(ctrl)
	txManager := mocks.NewMockManager(ctrl)

	service := translation.NewServiceTranslation(txManager, log, balanceUpdater, saverWallet, publisher)

	txManager.EXPECT().
		ExecuteInTransaction(gomock.Any(), "translation", gomock.Any()).
		DoAndReturn(func(ctx context.Context, name string, fn func(pgxdriver.QueryExecuter) error) error {
			return fn(nil)
		})

	balanceUpdater.EXPECT().
		DecreaseBalance(gomock.Any(), nil, fromUser, currency, amount).
		Return(int64(0), nil)

	balanceUpdater.EXPECT().
		IncreaseBalance(gomock.Any(), nil, toUser, currency, amount).
		Return(int64(0), storage.ErrWalletNotFound)

	saverWallet.EXPECT().
		CreateWallet(gomock.Any(), gomock.Any(), toUser, currency).
		Return(nil)

	balanceUpdater.EXPECT().
		IncreaseBalance(gomock.Any(), nil, toUser, currency, amount).
		Return(int64(amount), nil)

	err := service.Translation(context.Background(), fromUser, toUser, amount, currency)
	require.NoError(t, err)
}
