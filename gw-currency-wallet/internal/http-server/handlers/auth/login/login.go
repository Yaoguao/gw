package login

import (
	"context"
	"fmt"
	"gw-currency-wallet/internal/http-server/handlers"
	"gw-currency-wallet/internal/lib/helpers"
	"log/slog"
	"net/http"
)

type Auth interface {
	Login(ctx context.Context, username, password string) (string, error)
}

type request struct {
	Username string `json:"username" example:"user1"`
	Password string `json:"password" example:"password123"`
}
type response struct {
	Token string `json:"token,omitempty" example:"jwt-token"`
	Error string `json:"error,omitempty" example:"invalid credentials"`
}

// Login godoc
// @Summary      User login
// @Description  Authenticates user and returns JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      request   true  "Login credentials"
// @Success      200      {object}  response  "JWT token"
// @Failure      400      {object}  response  "Invalid request body"
// @Failure      401      {object}  response  "Invalid username or password"
// @Failure      500      {object}  response  "Internal server error"
// @Router       /login [post]
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

		token, err := auth.Login(r.Context(), req.Username, req.Password)

		if err != nil {
			log.Error("failed login", err.Error())
			handlers.ErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}

		err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"res": response{
			Token: token,
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

	return nil
}
