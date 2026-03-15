package auth

import (
	"context"
	"errors"
	"fmt"
	"gw-currency-wallet/internal/domain/models"
	"gw-currency-wallet/internal/lib/jwt"
	"gw-currency-wallet/internal/lib/logger/sl"
	"gw-currency-wallet/internal/storage"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserSaver interface {
	SaveUser(ctx context.Context, id uuid.UUID, username, email string, passHash []byte) error
}

type UserGetter interface {
	GetUser(ctx context.Context, username string) (*models.User, error)
}

type Auth struct {
	log *slog.Logger

	userSaver  UserSaver
	userGetter UserGetter

	tokenTTL time.Duration

	jwtSecret string
}

func New(log *slog.Logger, userSaver UserSaver, userGetter UserGetter, tokenTTL time.Duration, jwtSecret string) *Auth {
	return &Auth{
		log:        log,
		userSaver:  userSaver,
		userGetter: userGetter,
		tokenTTL:   tokenTTL,
		jwtSecret:  jwtSecret,
	}
}

func (a *Auth) Login(ctx context.Context, username, password string) (string, error) {
	const op = "auth.Login"

	user, err := a.userGetter.GetUser(ctx, username)

	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	token, err := jwt.NewToken(user, a.tokenTTL, a.jwtSecret)

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (a *Auth) Register(ctx context.Context, username, email, password string) (uuid.UUID, error) {
	const op = "auth.Register"

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		a.log.Error("failed to generate password hash", sl.Err(err))

		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	uid := uuid.New()

	err = a.userSaver.SaveUser(ctx, uid, username, email, passHash)

	if err != nil {
		a.log.Error("failed to save user", sl.Err(err))

		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return uid, nil
}
