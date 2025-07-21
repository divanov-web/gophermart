package main

import (
	"context"
	"github.com/divanov-web/gophermart/internal/accrual"
	"github.com/divanov-web/gophermart/internal/config"
	"github.com/divanov-web/gophermart/internal/handlers"
	"github.com/divanov-web/gophermart/internal/middleware"
	"github.com/divanov-web/gophermart/internal/repository"
	"github.com/divanov-web/gophermart/internal/service"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func main() {
	cfg := config.NewConfig()

	// создаём предустановленный регистратор zap
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	// делаем регистратор SugaredLogger
	sugar := logger.Sugar()
	middleware.SetLogger(sugar) // передаём логгер в middleware
	//сброс буфера логгера (добавлено про запас по урокам)
	defer func() {
		if err := logger.Sync(); err != nil {
			sugar.Errorw("Failed to sync logger", "error", err)
		}
	}()

	//context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = ctx

	gormDB, err := repository.InitDB(cfg.DatabaseDSN)
	if err != nil {
		sugar.Fatalw("failed to initialize database", "error", err)
	}

	userRepo := repository.NewUserRepository(gormDB)
	userService := service.NewUserService(userRepo)

	orderRepo := repository.NewOrderRepository(gormDB)
	orderService := service.NewOrderService(orderRepo, userRepo, sugar, cfg)

	h := handlers.NewHandler(userService, orderService, sugar, cfg)

	accrualClient := accrual.NewClient(cfg.AccrualAddress, sugar)
	orderService.StartOrderSenderWorker(ctx, 3*time.Second, accrualClient)
	orderService.StartAccrualUpdaterWorker(ctx, 5*time.Second, accrualClient)

	sugar.Infow(
		"Starting server",
		"addr", cfg.ServerAddress,
	)

	sugar.Infow("Config",
		"ServerAddress", cfg.ServerAddress,
		"DatabaseDSN", cfg.DatabaseDSN,
	)

	if err := http.ListenAndServe(cfg.ServerAddress, h.Router); err != nil {
		sugar.Fatalw("Server failed", "error", err)
	}

}
