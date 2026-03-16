package wallet_test

import (
	"context"
	"gw-currency-wallet/internal/domain/models"
	"gw-currency-wallet/internal/lib/logger/sl"
	"gw-currency-wallet/internal/services/wallet"
	"gw-currency-wallet/internal/services/wallet/mocks"
	"gw-currency-wallet/internal/storage"
	pgxdriver "gw-currency-wallet/pkg/pgx-driver"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestDeposit_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := sl.InitLogger("dev", os.Stdout)

	userID := uuid.New()
	amount := int64(1000)
	currency := "USD"

	balanceUpdater := mocks.NewMockBalanceUpdaterWallet(ctrl)
	saverWallet := mocks.NewMockSaverWallet(ctrl)
	saverTransaction := mocks.NewMockSaverTransaction(ctrl)
	txManager := mocks.NewMockManager(ctrl)
	getterWallet := mocks.NewMockGetterWallet(ctrl)

	service := wallet.NewServiceWallet(txManager, log, saverWallet, saverTransaction, getterWallet, balanceUpdater)

	txManager.EXPECT().
		ExecuteInTransaction(gomock.Any(), "deposit", gomock.Any()).
		DoAndReturn(func(ctx context.Context, name string, fn func(pgxdriver.QueryExecuter) error) error {
			return fn(nil)
		})

	balanceUpdater.EXPECT().
		IncreaseBalance(gomock.Any(), nil, userID, currency, amount).
		Return(amount, nil)

	saverTransaction.EXPECT().
		CreateTransaction(gomock.Any(), nil, gomock.Any()).
		Return(nil)

	err := service.Deposit(context.Background(), userID, currency, amount)
	require.NoError(t, err)
}

func TestDeposit_CreateWalletIfNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := sl.InitLogger("dev", os.Stdout)

	userID := uuid.New()
	amount := int64(1000)
	currency := "USD"

	balanceUpdater := mocks.NewMockBalanceUpdaterWallet(ctrl)
	saverWallet := mocks.NewMockSaverWallet(ctrl)
	saverTransaction := mocks.NewMockSaverTransaction(ctrl)
	txManager := mocks.NewMockManager(ctrl)
	getterWallet := mocks.NewMockGetterWallet(ctrl)

	service := wallet.NewServiceWallet(txManager, log, saverWallet, saverTransaction, getterWallet, balanceUpdater)

	txManager.EXPECT().
		ExecuteInTransaction(gomock.Any(), "deposit", gomock.Any()).
		DoAndReturn(func(ctx context.Context, name string, fn func(pgxdriver.QueryExecuter) error) error {
			return fn(nil)
		})

	balanceUpdater.EXPECT().
		IncreaseBalance(gomock.Any(), nil, userID, currency, amount).
		Return(int64(0), storage.ErrWalletNotFound)

	saverWallet.EXPECT().
		CreateWallet(gomock.Any(), gomock.Any(), userID, currency).
		Return(nil)

	balanceUpdater.EXPECT().
		IncreaseBalance(gomock.Any(), nil, userID, currency, amount).
		Return(amount, nil)

	saverTransaction.EXPECT().
		CreateTransaction(gomock.Any(), nil, gomock.Any()).
		Return(nil)

	err := service.Deposit(context.Background(), userID, currency, amount)
	require.NoError(t, err)
}

func TestWithdraw_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := sl.InitLogger("dev", os.Stdout)

	userID := uuid.New()
	amount := int64(500)
	currency := "USD"

	balanceUpdater := mocks.NewMockBalanceUpdaterWallet(ctrl)
	saverWallet := mocks.NewMockSaverWallet(ctrl)
	saverTransaction := mocks.NewMockSaverTransaction(ctrl)
	txManager := mocks.NewMockManager(ctrl)
	getterWallet := mocks.NewMockGetterWallet(ctrl)

	service := wallet.NewServiceWallet(txManager, log, saverWallet, saverTransaction, getterWallet, balanceUpdater)

	txManager.EXPECT().
		ExecuteInTransaction(gomock.Any(), "withdraw", gomock.Any()).
		DoAndReturn(func(ctx context.Context, name string, fn func(pgxdriver.QueryExecuter) error) error {
			return fn(nil)
		})

	balanceUpdater.EXPECT().
		DecreaseBalance(gomock.Any(), nil, userID, currency, amount).
		Return(amount, nil)

	saverTransaction.EXPECT().
		CreateTransaction(gomock.Any(), nil, gomock.Any()).
		Return(nil)

	err := service.Withdraw(context.Background(), userID, currency, amount)
	require.NoError(t, err)
}

func TestGetWalletsBalanceByUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := sl.InitLogger("dev", os.Stdout)

	userID := uuid.New()

	wallets := []models.Wallet{
		{Currency: "USD", Balance: 1000},
		{Currency: "EUR", Balance: 500},
	}

	getterWallet := mocks.NewMockGetterWallet(ctrl)
	service := wallet.NewServiceWallet(nil, log, nil, nil, getterWallet, nil)

	getterWallet.EXPECT().
		GetWalletsByUser(gomock.Any(), userID).
		Return(wallets, nil)

	balances, err := service.GetWalletsBalanceByUser(context.Background(), userID)
	require.NoError(t, err)
	require.Equal(t, int64(1000), balances["USD"])
	require.Equal(t, int64(500), balances["EUR"])
}

func TestGetWalletsBalanceByUser_InvalidUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := sl.InitLogger("dev", os.Stdout)

	service := wallet.NewServiceWallet(nil, log, nil, nil, nil, nil)

	balances, err := service.GetWalletsBalanceByUser(context.Background(), uuid.Nil)
	require.ErrorIs(t, err, wallet.ErrInvalidUserID)
	require.Nil(t, balances)
}
