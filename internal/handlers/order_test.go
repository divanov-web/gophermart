package handlers_test

import (
	"bytes"
	"context"
	"github.com/divanov-web/gophermart/internal/config"
	"github.com/divanov-web/gophermart/internal/handlers"
	"github.com/divanov-web/gophermart/internal/middleware"
	"github.com/divanov-web/gophermart/internal/mocks"
	"github.com/divanov-web/gophermart/internal/model"
	"github.com/divanov-web/gophermart/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOrderUploadHandler(t *testing.T) {
	type want struct {
		statusCode int
	}

	tests := []struct {
		name          string
		body          string
		mockSetup     func(*mocks.MockOrderRepo)
		expectedError error
		want          want
	}{
		{
			name: "valid order number, new",
			body: "79927398713", // валидный по Луну
			mockSetup: func(repo *mocks.MockOrderRepo) {
				repo.On("GetByNumber", mock.Anything, "79927398713").Return(nil, gorm.ErrRecordNotFound)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*model.Order")).Return(nil)
			},
			want: want{statusCode: http.StatusAccepted},
		},
		{
			name: "order already uploaded by same user",
			body: "79927398713",
			mockSetup: func(repo *mocks.MockOrderRepo) {
				repo.On("GetByNumber", mock.Anything, "79927398713").Return(&model.Order{Number: "79927398713", UserID: 42}, nil)
			},
			want: want{statusCode: http.StatusOK},
		},
		{
			name: "order already uploaded by another user",
			body: "79927398713",
			mockSetup: func(repo *mocks.MockOrderRepo) {
				repo.On("GetByNumber", mock.Anything, "79927398713").Return(&model.Order{Number: "79927398713", UserID: 999}, nil)
			},
			want: want{statusCode: http.StatusConflict},
		},
		{
			name:      "empty body",
			body:      "",
			mockSetup: func(repo *mocks.MockOrderRepo) {},
			want:      want{statusCode: http.StatusBadRequest},
		},
		{
			name:      "invalid order number",
			body:      "123456789", // невалидный по Луну
			mockSetup: func(repo *mocks.MockOrderRepo) {},
			want:      want{statusCode: http.StatusUnprocessableEntity},
		},
	}
	cfg := config.NewConfig()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderRepo := new(mocks.MockOrderRepo)
			if tt.mockSetup != nil {
				tt.mockSetup(orderRepo)
			}

			userRepo := new(mocks.MockUserRepo)
			logger := zaptest.NewLogger(t).Sugar()
			svc := service.NewOrderService(orderRepo, userRepo, logger, cfg)
			handler := handlers.NewOrderHandler(svc, logger)

			r := chi.NewRouter()
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ctx := context.WithValue(r.Context(), middleware.UserKey, int64(42))
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			})
			r.Post("/api/user/orders", handler.Upload)

			req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString(tt.body))
			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, req)

			assert.Equal(t, tt.want.statusCode, resp.Code)
		})
	}
}
