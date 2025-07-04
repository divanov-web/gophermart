package main

import (
	"context"
	"github.com/divanov-web/gophermart/internal/config"
	"github.com/divanov-web/gophermart/internal/handlers"
	"github.com/divanov-web/gophermart/internal/middleware"
	"github.com/divanov-web/gophermart/internal/service"
	"github.com/divanov-web/gophermart/internal/storage/pgstorage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
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

	pool, err := pgstorage.NewPool(ctx, cfg.DatabaseDSN)
	if err != nil {

		sugar.Fatalw("failed to initialize storage", "error", err)
	}
	store, err := pgstorage.NewStorage(ctx, pool)
	if err != nil {
		pool.Close()
		sugar.Fatalw("failed to initialize storage", "error", err)
	}

	urlService := service.NewURLService(ctx, store)
	h := handlers.NewHandler(urlService)

	r := chi.NewRouter()
	r.Use(middleware.WithGzip)
	r.Use(middleware.WithLogging) //логирование

	r.Post("/api/user/register", h.UserRegister)

	sugar.Infow(
		"Starting server",
		"addr", cfg.ServerAddress,
	)

	sugar.Infow("Config",
		"ServerAddress", cfg.ServerAddress,
		"DatabaseDSN", cfg.DatabaseDSN,
	)

	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		sugar.Fatalw("Server failed", "error", err)
	}

}
