package service_test

import (
	"context"
	"errors"
	"github.com/divanov-web/gophermart/internal/mocks"
	"testing"

	"github.com/divanov-web/gophermart/internal/model"
	"github.com/divanov-web/gophermart/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestRegisterUser_Success(t *testing.T) {
	repo := new(mocks.MockUserRepo)
	svc := service.NewUserService(repo)

	ctx := context.Background()
	login := "testuser1"
	password := "password123"

	repo.On("GetUserByLogin", ctx, login).Return(nil, errors.New("not found"))
	repo.On("CreateUser", ctx, mock.AnythingOfType("*model.User")).Return(&model.User{ID: 1, Login: login}, nil)

	user, err := svc.Register(ctx, login, password)

	assert.NoError(t, err)
	assert.Equal(t, login, user.Login)
}

func TestRegisterUser_LoginTaken(t *testing.T) {
	repo := new(mocks.MockUserRepo)
	svc := service.NewUserService(repo)

	ctx := context.Background()
	login := "testuser1"
	password := "password123"

	repo.On("GetUserByLogin", ctx, login).Return(&model.User{Login: login}, nil)

	user, err := svc.Register(ctx, login, password)

	assert.ErrorIs(t, err, service.ErrLoginTaken)
	assert.Nil(t, user)
}

func TestLogin_Success(t *testing.T) {
	repo := new(mocks.MockUserRepo)
	svc := service.NewUserService(repo)

	ctx := context.Background()
	login := "testuser1"
	password := "password123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	repo.On("GetUserByLogin", ctx, login).Return(&model.User{Login: login, Password: string(hash)}, nil)

	user, err := svc.Login(ctx, login, password)

	assert.NoError(t, err)
	assert.Equal(t, login, user.Login)
}

func TestLogin_InvalidPassword(t *testing.T) {
	repo := new(mocks.MockUserRepo)
	svc := service.NewUserService(repo)

	ctx := context.Background()
	login := "testuser1"
	hash, _ := bcrypt.GenerateFromPassword([]byte("otherpass"), bcrypt.DefaultCost)

	repo.On("GetUserByLogin", ctx, login).Return(&model.User{Login: login, Password: string(hash)}, nil)

	user, err := svc.Login(ctx, login, "wrongpassword")

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestGetUserBalance(t *testing.T) {
	repo := new(mocks.MockUserRepo)
	svc := service.NewUserService(repo)

	ctx := context.Background()
	userID := int64(42)

	repo.On("GetBalance", ctx, userID).Return(100.5, 40.0, nil)

	resp, err := svc.GetUserBalance(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, 100.5, resp.Current)
	assert.Equal(t, 40.0, resp.Withdrawn)
}
