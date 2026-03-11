package main

import (
	"context"
	"encoding/json"
	"gw-exchanger/internal/config"
	"gw-exchanger/internal/lib/logger/sl"
	"gw-exchanger/internal/storage/postgres"
	pgxdriver "gw-exchanger/pkg/pgx-driver"
	"net/http"
	"os"
)

type ApiResponse struct {
	Source string             `json:"source"`
	Quotes map[string]float64 `json:"quotes"`
}

func main() {
	ctx := context.Background()

	cfg := config.MustLoad()

	logger := sl.InitLogger(cfg.Env, os.Stdout)

	logger.Debug("CONFIG", cfg)

	pCfg := cfg.StorageConfig.Postgres

	// INIT CONNECTION DB

	storage, err := pgxdriver.New(
		pCfg.DSN,
		logger,
		pgxdriver.MaxPoolSize(pCfg.MaxOpenConns),
		pgxdriver.MinConns(pCfg.MaxIdleConns),
		pgxdriver.MaxConnIdleTime(pCfg.MaxIdleTime))

	if err != nil {
		panic(err)
	}
	defer storage.Close()
	// INIT REPOSITORY

	repo := postgres.NewExchangeRatesRepository(logger, storage)

	rates, err := fetchRates()
	if err != nil {
		logger.Error("fetch error", "err", err.Error())
		return
	}

	pairs := buildPairs(rates)

	err = repo.SaveBatch(ctx, pairs)
	if err != nil {
		logger.Error("save batch failed", "err", err.Error())
	}

	logger.Info("done")

}

func fetchRates() (map[string]float64, error) {
	url := "https://api.exchangerate.host/live?access_key=aad6e0721250ab263318414d8b65ffaf&base=USD"

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data ApiResponse

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	rates := make(map[string]float64)

	for k, v := range data.Quotes {
		currency := k[3:]
		rates[currency] = v
	}

	rates["USD"] = 1

	return rates, nil
}

func buildPairs(rates map[string]float64) map[string]map[string]float64 {

	result := make(map[string]map[string]float64)

	for from, fromRate := range rates {

		result[from] = make(map[string]float64)

		for to, toRate := range rates {

			if from == to {
				continue
			}

			rate := toRate / fromRate

			result[from][to] = rate
		}
	}

	return result
}
