package service

import (
	"context"
	"errors"
	"github.com/divanov-web/gophermart/internal/model"
	"github.com/divanov-web/gophermart/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrOrderExists             = errors.New("order already uploaded")
	ErrOrderOwnedByOther       = errors.New("order already uploaded by another user")
	ErrInvalidOrderNumber      = errors.New("wrong order number")
	ErrOrderOwnedByAnotherUser = errors.New("wrong order owned by another user")
)

type OrderService struct {
	repo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

// UploadOrder загружает новый заказ
func (s *OrderService) UploadOrder(ctx context.Context, userID int64, number string) error {
	order, err := s.repo.GetByNumber(ctx, number)
	if err == nil {
		if order.UserID == userID {
			return ErrOrderExists
		}
		return ErrOrderOwnedByOther
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// если заказ не найден, создаём новый
	newOrder := &model.Order{
		Number: number,
		UserID: userID,
		Status: model.OrderStatusNew,
	}
	return s.repo.Create(ctx, newOrder)
}

func (s *OrderService) GetOrdersByUser(ctx context.Context, userID int64) ([]model.Order, error) {
	return s.repo.GetByUserID(ctx, userID)
}
