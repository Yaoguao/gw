package register

import (
	"context"
	"fmt"
	"gw-currency-wallet/internal/http-server/handlers"
	"gw-currency-wallet/internal/lib/helpers"
	"log/slog"
	"net/http"
	"regexp"

	"github.com/google/uuid"
)

var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

type Auth interface {
	Register(ctx context.Context, username, email, password string) (uuid.UUID, error)
}

type request struct {
	Username string `json:"username" example:"user1"`
	Password string `json:"password" example:"password123"`
	Email    string `json:"email" example:"user1@example.com"`
}

type response struct {
	Message string `json:"message,omitempty" example:"user registered successfully"`
	Error   string `json:"error,omitempty" example:"email already exists"`
}

// Register godoc
// @Summary      Register new user
// @Description  Creates a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      request   true  "User registration data"
// @Success      200      {object}  response  "User registered successfully"
// @Failure      400      {object}  response  "Invalid request body"
// @Failure      401      {object}  response  "Validation error"
// @Failure      500      {object}  response  "Internal server error"
// @Router       /register [post]
func New(log *slog.Logger, auth Auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req request

		err := helpers.ReadJSON(w, r, &req)

		if err != nil {
			log.Error("failed read json")
			handlers.BadRequestResponse(w, r, fmt.Errorf("failed read json"))
			return
		}

		err = validRequest(req)

		if err != nil {
			handlers.ErrorResponse(w, r, http.StatusUnauthorized, err.Error())
			return
		}

		_, err = auth.Register(r.Context(), req.Username, req.Email, req.Password)

		if err != nil {
			log.Error("failed register", err.Error())
			handlers.ErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}

		err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"res": response{
			Message: "user registered successfully",
		}}, w.Header())

		if err != nil {
			handlers.BadRequestResponse(w, r, err)
		}
	}
}

func validRequest(req request) error {
	if len(req.Username) == 0 {
		return fmt.Errorf("username is empty")
	}

	if len(req.Password) <= 4 || len(req.Password) >= 72 {
		return fmt.Errorf("password: must be at least 8 bytes long or must not be more than 72 bytes long")
	}

	if !EmailRX.MatchString(req.Email) {
		return fmt.Errorf("must be valid email address")
	}

	return nil
}
