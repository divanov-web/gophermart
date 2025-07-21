package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/divanov-web/gophermart/internal/config"
	"github.com/divanov-web/gophermart/internal/mocks"
	"github.com/divanov-web/gophermart/internal/model"
	"github.com/divanov-web/gophermart/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testSecret = "test-secret"

func TestRegisterHandler_Success(t *testing.T) {
	repo := new(mocks.MockUserRepo)
	svc := service.NewUserService(repo)
	logger := zap.NewNop().Sugar()
	cfg := &config.Config{AuthSecret: testSecret}
	handler := NewUserHandler(svc, logger, cfg)

	login := "testuser"
	password := "123456"
	requestBody, _ := json.Marshal(RegisterRequest{Login: login, Password: password})

	repo.On("GetUserByLogin", mock.Anything, login).Return(nil, errors.New("not found"))
	repo.On("CreateUser", mock.Anything, mock.Anything).Return(&model.User{ID: 1, Login: login}, nil)

	r := chi.NewRouter()
	r.Post("/api/user/register", handler.Register)

	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	trw := httptest.NewRecorder()

	r.ServeHTTP(trw, req)
	res := trw.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	cookieFound := false
	for _, c := range res.Cookies() {
		if c.Name == "auth_token" {
			cookieFound = true
			break
		}
	}
	assert.True(t, cookieFound, "auth_token cookie not set")
}

func TestRegisterHandler_LoginTaken(t *testing.T) {
	repo := new(mocks.MockUserRepo)
	svc := service.NewUserService(repo)
	logger := zap.NewNop().Sugar()
	cfg := &config.Config{AuthSecret: testSecret}
	handler := NewUserHandler(svc, logger, cfg)

	login := "testuser"
	password := "123456"
	requestBody, _ := json.Marshal(RegisterRequest{Login: login, Password: password})

	repo.On("GetUserByLogin", mock.Anything, login).Return(&model.User{Login: login}, nil)

	r := chi.NewRouter()
	r.Post("/api/user/register", handler.Register)

	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	trw := httptest.NewRecorder()

	r.ServeHTTP(trw, req)
	res := trw.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusConflict, res.StatusCode)
}

func TestLoginHandler_Success(t *testing.T) {
	repo := new(mocks.MockUserRepo)
	svc := service.NewUserService(repo)
	logger := zap.NewNop().Sugar()
	cfg := &config.Config{AuthSecret: testSecret}
	handler := NewUserHandler(svc, logger, cfg)

	login := "testuser"
	password := "123456"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	repo.On("GetUserByLogin", mock.Anything, login).Return(&model.User{ID: 1, Login: login, Password: string(hash)}, nil)

	requestBody, _ := json.Marshal(LoginRequest{Login: login, Password: password})
	r := chi.NewRouter()
	r.Post("/api/user/login", handler.Login)

	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	trw := httptest.NewRecorder()

	r.ServeHTTP(trw, req)
	res := trw.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	repo := new(mocks.MockUserRepo)
	svc := service.NewUserService(repo)
	logger := zap.NewNop().Sugar()
	cfg := &config.Config{AuthSecret: testSecret}
	handler := NewUserHandler(svc, logger, cfg)

	login := "testuser"
	correctHash, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)
	repo.On("GetUserByLogin", mock.Anything, login).Return(&model.User{Login: login, Password: string(correctHash)}, nil)

	requestBody, _ := json.Marshal(LoginRequest{Login: login, Password: "wrongpass"})
	r := chi.NewRouter()
	r.Post("/api/user/login", handler.Login)

	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	trw := httptest.NewRecorder()

	r.ServeHTTP(trw, req)
	res := trw.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}
