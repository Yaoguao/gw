package get

import (
	"context"
	"gw-currency-wallet/internal/http-server/handlers"
	"gw-currency-wallet/internal/lib/helpers"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ERGetter interface {
	GetExchangeRates(ctx context.Context, base string) (map[string]float64, error)
}

type response struct {
	Rates map[string]float64 `json:"rates,omitempty"`
	Error string             `json:"error,omitempty" example:"invalid base currency"`
}

// GetExchangeRates godoc
// @Summary      Get exchange rates
// @Description  Returns exchange rates for the specified base currency
// @Tags         exchange
// @Produce      json
// @Param        BASE   path      string   true  "Base currency code (ISO 4217)" example(USD)
// @Success      200    {object}  response "Exchange rates"
// @Failure      400    {object}  response "Invalid base currency"
// @Failure      500    {object}  response "Internal server error"
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Enter your JWT token as: Bearer <token>
// @Router       /exchange/rates/{BASE} [get]
func New(log *slog.Logger, er ERGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		base := chi.URLParam(r, "BASE")
		if len(base) != 3 {
			handlers.ErrorResponse(w, r, http.StatusBadRequest, "no correct format base")
			return
		}

		eRates, err := er.GetExchangeRates(r.Context(), base)

		if err != nil {
			log.Error("error get exchange rates", err.Error())
			handlers.ErrorResponse(w, r, http.StatusInternalServerError, "error get exchange rates")
			return
		}

		err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"res": response{
			Rates: eRates,
		}}, w.Header())

		if err != nil {
			handlers.BadRequestResponse(w, r, err)
		}
	}
}
