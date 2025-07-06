package repository

import (
	"context"
	"errors"

	"github.com/divanov-web/gophermart/internal/model"
	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByNumber(ctx context.Context, number string) (*model.Order, error)
}

type orderRepo struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepo{db: db}
}

func (r *orderRepo) Create(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *orderRepo) GetByNumber(ctx context.Context, number string) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).Where("number = ?", number).First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound // это корректное поведение — заказ не найден
	}
	return &order, err
}
