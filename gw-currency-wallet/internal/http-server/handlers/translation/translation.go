package translation

import (
	"context"
	"fmt"
	"gw-currency-wallet/internal/http-server/handlers"
	"gw-currency-wallet/internal/lib/helpers"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type Translatable interface {
	Translation(ctx context.Context, fromUser uuid.UUID, toUser uuid.UUID, amount int64, currency string) error
}

type request struct {
	ToUserID string  `json:"to_user_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

func New(log *slog.Logger, translatable Translatable) http.HandlerFunc {
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

		toUID, err := uuid.Parse(req.ToUserID)

		if err != nil {
			log.Error("error parse uid", err)
			handlers.BadRequestResponse(w, r, fmt.Errorf("failed parse uuid"))
			return
		}

		err = translatable.Translation(
			r.Context(),
			uid,
			toUID,
			helpers.ParseAmount(req.Amount),
			req.Currency,
		)

		if err != nil {
			log.Error("error parse uid", err)
			handlers.ErrorResponse(w, r, http.StatusInternalServerError, "error translation operation")
			return
		}

		err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "success translation operation"}, w.Header())

		if err != nil {
			handlers.BadRequestResponse(w, r, err)
		}
	}
}
