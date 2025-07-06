package handlers

import (
	"github.com/divanov-web/gophermart/internal/config"
	"github.com/divanov-web/gophermart/internal/middleware"
	"github.com/divanov-web/gophermart/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Handler struct {
	Router chi.Router
}

// NewHandler разводящий для хендлеров
func NewHandler(userService *service.UserService, logger *zap.SugaredLogger, config *config.Config) *Handler {
	r := chi.NewRouter()

	r.Use(middleware.WithGzip)
	r.Use(middleware.WithLogging) //логирование
	r.Use(middleware.WithAuth(config.AuthSecret))

	userHandler := NewUserHandler(userService, logger, config)
	r.Post("/api/user/register", userHandler.Register)
	r.Post("/api/user/test", userHandler.Test)
	r.Post("/api/user/login", userHandler.Login)

	return &Handler{Router: r}
}
