package auth_test

import (
	"context"
	"errors"
	"gw-currency-wallet/internal/domain/models"
	"gw-currency-wallet/internal/services/auth"
	"gw-currency-wallet/internal/services/auth/mocks"
	"gw-currency-wallet/internal/storage"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestLogin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userGetter := mocks.NewMockUserGetter(ctrl)
	userSaver := mocks.NewMockUserSaver(ctrl)

	password := "password123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &models.User{
		ID:       uuid.New(),
		Username: "test",
		PassHash: hash,
	}

	userGetter.
		EXPECT().
		GetUser(gomock.Any(), "test").
		Return(user, nil)

	service := auth.New(
		slog.Default(),
		userSaver,
		userGetter,
		time.Hour,
		"secret",
	)

	token, err := service.Login(context.Background(), "test", password)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestLogin_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userGetter := mocks.NewMockUserGetter(ctrl)
	userSaver := mocks.NewMockUserSaver(ctrl)

	userGetter.
		EXPECT().
		GetUser(gomock.Any(), "test").
		Return(nil, storage.ErrUserNotFound)

	service := auth.New(
		slog.Default(),
		userSaver,
		userGetter,
		time.Hour,
		"secret",
	)

	_, err := service.Login(context.Background(), "test", "password")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, auth.ErrInvalidCredentials))
}

func TestLogin_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userGetter := mocks.NewMockUserGetter(ctrl)
	userSaver := mocks.NewMockUserSaver(ctrl)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)

	user := &models.User{
		ID:       uuid.New(),
		Username: "test",
		PassHash: hash,
	}

	userGetter.
		EXPECT().
		GetUser(gomock.Any(), "test").
		Return(user, nil)

	service := auth.New(
		slog.Default(),
		userSaver,
		userGetter,
		time.Hour,
		"secret",
	)

	_, err := service.Login(context.Background(), "test", "wrong")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, auth.ErrInvalidCredentials))
}

func TestRegister_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userGetter := mocks.NewMockUserGetter(ctrl)
	userSaver := mocks.NewMockUserSaver(ctrl)

	userSaver.
		EXPECT().
		SaveUser(
			gomock.Any(),
			gomock.Any(),
			"test",
			"test@mail.com",
			gomock.Any(),
		).
		Return(nil)

	service := auth.New(
		slog.Default(),
		userSaver,
		userGetter,
		time.Hour,
		"secret",
	)

	id, err := service.Register(
		context.Background(),
		"test",
		"test@mail.com",
		"password",
	)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)
}
