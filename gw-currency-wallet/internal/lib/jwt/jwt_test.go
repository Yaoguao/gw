package jwt

import (
	"gw-currency-wallet/internal/domain/models"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestNewToken(t *testing.T) {
	uid, _ := uuid.NewUUID()
	user := models.User{
		ID:    uid,
		Email: "test@example.com",
	}

	secret := "supersecret"

	duration := time.Hour

	tokenString, err := NewToken(&user, duration, secret)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokenString == "" {
		t.Fatal("expected token, got empty string")
	}

	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	if !parsedToken.Valid {
		t.Fatal("token is not valid")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("invalid claims type")
	}

	if claims["uid"] != user.ID.String() {
		t.Errorf("expected uid %d, got %v", user.ID, claims["uid"])
	}

	if claims["email"] != user.Email {
		t.Errorf("expected email %s, got %v", user.Email, claims["email"])
	}

	if claims["username"] != user.Username {
		t.Errorf("expected app_id %s, got %v", user.Username, claims["username"])
	}
}
