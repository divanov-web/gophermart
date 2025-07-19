package service

import (
	"context"
	"errors"
	"github.com/divanov-web/gophermart/internal/accrual"
	"github.com/divanov-web/gophermart/internal/model"
	"github.com/divanov-web/gophermart/internal/repository"
	"github.com/divanov-web/gophermart/internal/utils"
	"gorm.io/gorm"
	"time"
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
	return &OrderService{
		repo: repo,
	}
}

// UploadOrder загружает новый заказ
func (s *OrderService) UploadOrder(ctx context.Context, userID int64, number string) error {
	if !utils.IsValidLuhn(number) {
		return ErrInvalidOrderNumber
	}

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

// StartOrderSenderWorker Создаёт горутину, отправляет заказы в Accrual (только для локального сервера)
func (s *OrderService) StartOrderSenderWorker(ctx context.Context, interval time.Duration, client *accrual.Client) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.processNewOrders(ctx, client)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// processNewOrders Отправляет заказы в Accrual и меняет статус с NEW на PROCESSING
func (s *OrderService) processNewOrders(ctx context.Context, client *accrual.Client) {
	orders, err := s.repo.GetByStatus(ctx, model.OrderStatusNew)
	if err != nil {
		// todo лог ошибки
		return
	}

	for _, order := range orders {
		err := client.SendOrder(order.Number)
		if err != nil {
			// todo лог ошибки
			continue
		}

		// Обновляем статус заказа на PROCESSING
		order.Status = model.OrderStatusProcessing
		_ = s.repo.Update(ctx, &order)
	}
}

// StartAccrualUpdaterWorker Создаёт горутину, периодически проверяет статус заказа в Accrual
func (s *OrderService) StartAccrualUpdaterWorker(ctx context.Context, interval time.Duration, client *accrual.Client) {
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.updateProcessingOrders(ctx, client)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// updateProcessingOrders Проверяет заказы в Accrual и меняет статус с PROCESSING на INVALID или PROCESSED
func (s *OrderService) updateProcessingOrders(ctx context.Context, client *accrual.Client) {
	orders, err := s.repo.GetByStatus(ctx, model.OrderStatusProcessing)
	if err != nil {
		// TODO: лог ошибки
		return
	}

	for _, order := range orders {
		resp, err := client.GetOrderInfo(order.Number)
		if err != nil || resp == nil {
			// TODO: лог ошибки или пропуск необработанного заказа
			continue
		}

		switch resp.Status {
		case "REGISTERED", "PROCESSING":
			// оставим без изменений
			continue
		case "PROCESSED":
			order.Status = model.OrderStatusProcessed
			order.Accrual = resp.Accrual
			_ = s.repo.Update(ctx, &order)
		case "INVALID":
			order.Status = model.OrderStatusInvalid
			_ = s.repo.Update(ctx, &order)
		default:
			// TODO: лог неизвестного статуса
			continue
		}
	}
}
