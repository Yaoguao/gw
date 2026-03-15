package balance

import (
	"context"
	"gw-currency-wallet/internal/http-server/handlers"
	"gw-currency-wallet/internal/lib/helpers"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type WalletGetter interface {
	GetWalletsBalanceByUser(ctx context.Context, userID uuid.UUID) (map[string]int64, error)
}

type response struct {
	Balance map[string]float64 `json:"balance,omitempty"`
	Error   string             `json:"error,omitempty" example:"failed get balance"`
}

// GetBalance godoc
// @Summary      Get user wallet balances
// @Description  Returns balances for all currencies belonging to the authenticated user
// @Tags         wallet
// @Produce      json
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Enter your JWT token as: Bearer <token>
// @Success      200  {object}  response  "User balances"
// @Failure      400  {object}  response  "Invalid user id"
// @Failure      401  {object}  response  "Unauthorized"
// @Failure      500  {object}  response  "Internal server error"
// @Router       /balance [get]
func New(log *slog.Logger, wallet WalletGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		uid, err := uuid.Parse(r.Context().Value("uid").(string))

		if err != nil {
			log.Error("error parse uid", err.Error())
			handlers.ErrorResponse(w, r, http.StatusBadRequest, "error parse uid")
			return
		}

		balance, err := wallet.GetWalletsBalanceByUser(r.Context(), uid)

		if err != nil {
			log.Error("error get balance", err.Error())
			handlers.ErrorResponse(w, r, http.StatusInternalServerError, "error get balance")
			return
		}

		err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"balance": response{
			Balance: helpers.FormatAmounts(balance),
		}}, w.Header())

		if err != nil {
			handlers.BadRequestResponse(w, r, err)
		}

	}
}
