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
func NewHandler(
	userService *service.UserService,
	orderService *service.OrderService,
	logger *zap.SugaredLogger,
	config *config.Config,
) *Handler {
	r := chi.NewRouter()

	r.Use(middleware.WithGzip)
	r.Use(middleware.WithLogging)
	r.Use(middleware.WithAuth(config.AuthSecret))

	// Handlers
	userHandler := NewUserHandler(userService, logger, config)
	orderHandler := NewOrderHandler(orderService, logger)
	balanceHandler := NewBalanceHandler(orderService, userService, logger)

	// User routes
	r.Post("/api/user/register", userHandler.Register)
	r.Post("/api/user/login", userHandler.Login)
	r.Post("/api/user/test", userHandler.Test)

	// Order routes
	r.Post("/api/user/orders", orderHandler.Upload)
	r.Get("/api/user/orders", orderHandler.GetUserOrders)

	// Withdraw routes
	r.Post("/api/user/balance/withdraw", balanceHandler.Withdraw)
	r.Get("/api/user/balance", balanceHandler.GetBalance)
	r.Get("/api/user/withdrawals", balanceHandler.GetWithdrawals)

	return &Handler{Router: r}
}
