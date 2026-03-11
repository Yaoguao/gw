package models

import "time"

type ExchangeRates struct {
	FromCurrency string    `json:"from_currency"`
	ToCurrency   string    `json:"to_currency"`
	Rate         float64   `json:"rate"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ExchangeRatesResponse struct {
	BaseCurrency string
	Rates        map[string]float64
}
