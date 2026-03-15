package app

import (
	httpapp "gw-currency-wallet/internal/app/http"
	"gw-currency-wallet/internal/config"
	"gw-currency-wallet/internal/http-server/handlers/auth/login"
	"gw-currency-wallet/internal/http-server/handlers/auth/register"
	"gw-currency-wallet/internal/http-server/handlers/exchanger/exchange"
	"gw-currency-wallet/internal/http-server/handlers/exchanger/get"
	"gw-currency-wallet/internal/http-server/handlers/wallet/balance"
	"gw-currency-wallet/internal/http-server/handlers/wallet/deposit"
	"gw-currency-wallet/internal/http-server/handlers/wallet/withdraw"
	authmiddleware "gw-currency-wallet/internal/http-server/middleware/auth"
	"gw-currency-wallet/internal/http-server/middleware/logger"
	"gw-currency-wallet/internal/services/auth"
	"gw-currency-wallet/internal/services/exchanger"
	"gw-currency-wallet/internal/services/wallet"
	exchangerstorage "gw-currency-wallet/internal/storage/exchanger"
	"gw-currency-wallet/internal/storage/postgres"
	pgxdriver "gw-currency-wallet/pkg/pgx-driver"
	"gw-currency-wallet/pkg/pgx-driver/transaction"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

type App struct {
	HTTPServer *httpapp.App

	Postgres *pgxdriver.Postgres
}

func New(log *slog.Logger, cfg *config.Config) *App {
	pCfg := cfg.StorageConfig.Postgres

	storage, err := pgxdriver.New(
		pCfg.DSN,
		log,
		pgxdriver.MaxPoolSize(pCfg.MaxOpenConns),
		pgxdriver.MinConns(pCfg.MaxIdleConns),
		pgxdriver.MaxConnIdleTime(pCfg.MaxIdleTime))

	if err != nil {
		panic(err)
	}

	txManger, err := transaction.NewManager(storage, log)
	if err != nil {
		panic(err)
	}

	userRepo := postgres.NewUserRepository(log, storage)
	walletRepo := postgres.NewWalletRepository(log, storage)
	exchangeRepo := postgres.NewExchangeRepository(log, storage)
	transactionRepo := postgres.NewTransactionRepository(log, storage)
	exchangeClient, err := exchangerstorage.NewClient(cfg.GwExchange.GRPC.Addr)
	if err != nil {
		log.Error("panic error", err.Error())
		panic(err)
	}

	authService := auth.New(log, userRepo, userRepo, cfg.JWT.TokenTTL, cfg.JWT.Secret)
	exchangerService := exchanger.NewServiceExchanger(txManger, log, exchangeClient, exchangeRepo, walletRepo, walletRepo, cfg.CacheTTL)
	walletService := wallet.NewServiceWallet(txManger, log, walletRepo, transactionRepo, walletRepo, walletRepo)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(logger.NewLoggerMiddleware(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Handle("/swagger/*", httpSwagger.WrapHandler)

	router.Route("/api/v1", func(r chi.Router) {

		r.Post("/login", login.New(log, authService))
		r.Post("/register", register.New(log, authService))

		r.Group(func(r chi.Router) {
			r.Use(authmiddleware.MiddlewareAuth(cfg.JWT.Secret))

			r.Get("/balance", balance.New(log, walletService))
			r.Post("/wallet/deposit", deposit.New(log, walletService, walletService))
			r.Post("/wallet/withdraw", withdraw.New(log, walletService, walletService))
			r.Get("/exchange/rates/{BASE}", get.New(log, exchangerService))
			r.Post("/exchange", exchange.New(log, exchangerService, walletService))
		})
	})

	server := httpapp.New(
		log,
		cfg.HTTPServer.Address,
		cfg.HTTPServer.Timeout,
		cfg.HTTPServer.Timeout,
		cfg.HTTPServer.IdleTimeout,
		router,
	)

	return &App{
		HTTPServer: server,
		Postgres:   storage,
	}
}
