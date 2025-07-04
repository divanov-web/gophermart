package service

import (
	"context"
	"github.com/divanov-web/gophermart/internal/storage"
)

type URLService struct {
	Repo storage.Storage
}

func NewURLService(ctx context.Context, repo storage.Storage) *URLService {
	svc := &URLService{
		Repo: repo,
	}

	return svc
}
