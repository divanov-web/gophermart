package service

import (
	"context"
	"github.com/divanov-web/gophermart/internal/model"
	"github.com/divanov-web/gophermart/internal/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Register регистрирует нового пользователя
func (s *UserService) Register(ctx context.Context, login, password string) error {
	// TODO: хешировать пароль, проверить уникальность логина
	return nil
}

// Login проверяет логин/пароль и возвращает пользователя
func (s *UserService) Login(ctx context.Context, login, password string) (*model.User, error) {
	// TODO: сравнить хеш пароля
	return nil, nil
}
