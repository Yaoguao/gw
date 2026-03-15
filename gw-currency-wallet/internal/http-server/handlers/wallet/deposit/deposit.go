package deposit

import (
	"context"
	"fmt"
	"gw-currency-wallet/internal/http-server/handlers"
	"gw-currency-wallet/internal/lib/helpers"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type WalletGetter interface {
	GetWalletsBalanceByUser(ctx context.Context, userID uuid.UUID) (map[string]int64, error)
}

type Depositable interface {
	Deposit(ctx context.Context, userID uuid.UUID, currency string, amount int64) error
}

type request struct {
	Amount   float64 `json:"amount" example:"100.50"`
	Currency string  `json:"currency" example:"USD"`
}

type response struct {
	Message string             `json:"message,omitempty" example:"Account topped up successfully"`
	Balance map[string]float64 `json:"balance,omitempty"`
	Error   string             `json:"error,omitempty" example:"invalid request"`
}

// Deposit godoc
// @Summary      Deposit funds to wallet
// @Description  Adds money to the authenticated user's wallet in the specified currency
// @Tags         wallet
// @Accept       json
// @Produce      json
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Enter your JWT token as: Bearer <token>
// @Param        request  body      request  true  "Deposit request"
// @Success      200      {object}  response "Deposit successful"
// @Failure      400      {object}  response "Invalid request"
// @Failure      401      {object}  response "Unauthorized"
// @Failure      500      {object}  response "Internal server error"
// @Router       /wallet/deposit [post]
func New(log *slog.Logger, wallet WalletGetter, depositable Depositable) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := uuid.Parse(r.Context().Value("uid").(string))

		if err != nil {
			log.Error("error parse uid", err.Error())
			handlers.ErrorResponse(w, r, http.StatusBadRequest, "error parse uid")
		}

		var req request

		err = helpers.ReadJSON(w, r, &req)

		if err != nil {
			log.Error("failed read json")
			handlers.BadRequestResponse(w, r, fmt.Errorf("failed read json"))
			return
		}

		if err := validateRequest(req); err != nil {
			handlers.BadRequestResponse(w, r, err)
			return
		}

		amount := helpers.ParseAmount(req.Amount)

		err = depositable.Deposit(r.Context(), uid, req.Currency, amount)

		if err != nil {
			log.Error("error deposit balance", err.Error())
			handlers.ErrorResponse(w, r, http.StatusInternalServerError, err.Error())
			return
		}

		balance, err := wallet.GetWalletsBalanceByUser(r.Context(), uid)

		if err != nil {
			log.Error("error get balance", err.Error())
			handlers.ErrorResponse(w, r, http.StatusInternalServerError, err.Error())
			return
		}

		err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"req": response{
			Message: "Account topped up successfully",
			Balance: helpers.FormatAmounts(balance),
		}}, w.Header())

		if err != nil {
			handlers.BadRequestResponse(w, r, err)
		}

	}
}

func validateRequest(req request) error {

	if req.Amount <= 0 {
		return fmt.Errorf("negative amount")
	}

	if len(req.Currency) != 3 {
		return fmt.Errorf("no correct format currency")
	}

	return nil
}
