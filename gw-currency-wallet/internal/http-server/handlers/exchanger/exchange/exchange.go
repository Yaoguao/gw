package exchange

import (
	"context"
	"fmt"
	"gw-currency-wallet/internal/http-server/handlers"
	"gw-currency-wallet/internal/lib/helpers"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type Exchangeable interface {
	Exchange(ctx context.Context, userID uuid.UUID, from, to string, amount int64) (int64, error)
}

type WalletGetter interface {
	GetWalletsBalanceByUser(ctx context.Context, userID uuid.UUID) (map[string]int64, error)
}

type request struct {
	FromCurrency string  `json:"from_currency" example:"USD"`
	ToCurrency   string  `json:"to_currency" example:"EUR"`
	Amount       float64 `json:"amount" example:"100.50"`
}

type response struct {
	Message         string             `json:"message,omitempty" example:"exchange successful"`
	ExchangedAmount float64            `json:"exchanged_amount" example:"92.35"`
	Balance         map[string]float64 `json:"balance,omitempty"`
	Error           string             `json:"error,omitempty" example:"insufficient funds"`
}

// Exchange godoc
// @Summary      Exchange currency
// @Description  Converts funds from one currency to another for the authenticated user
// @Tags         wallet
// @Accept       json
// @Produce      json
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Enter your JWT token as: Bearer <token>
// @Param        request  body      request   true  "Currency exchange request"
// @Success      200      {object}  response  "Exchange successful"
// @Failure      400      {object}  response  "Invalid request or user id"
// @Failure      401      {object}  response  "Unauthorized"
// @Failure      500      {object}  response  "Internal server error"
// @Router       /exchange [post]
func New(log *slog.Logger, exchangeable Exchangeable, wallet WalletGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := uuid.Parse(r.Context().Value("uid").(string))

		if err != nil {
			log.Error("error parse uid", err.Error())
			handlers.ErrorResponse(w, r, http.StatusBadRequest, "error parse uid")
			return
		}

		var req request

		err = helpers.ReadJSON(w, r, &req)

		if err != nil {
			log.Error("failed read json")
			handlers.BadRequestResponse(w, r, fmt.Errorf("failed read json"))
			return
		}

		exAmount, err := exchangeable.Exchange(
			r.Context(),
			uid,
			req.FromCurrency,
			req.ToCurrency,
			helpers.ParseAmount(req.Amount),
		)

		if err != nil {
			log.Error("error exchange currency", err.Error())
			handlers.ErrorResponse(w, r, http.StatusInternalServerError, "error exchange currency")
			return
		}

		balance, err := wallet.GetWalletsBalanceByUser(r.Context(), uid)

		if err != nil {
			log.Error("error get balance", err.Error())
			handlers.ErrorResponse(w, r, http.StatusInternalServerError, err.Error())
			return
		}

		err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"req": response{
			Message:         "Account topped up successfully",
			ExchangedAmount: helpers.FormatAmount(exAmount),
			Balance:         helpers.FormatAmounts(balance),
		}}, w.Header())

		if err != nil {
			handlers.BadRequestResponse(w, r, err)
		}

	}
}
