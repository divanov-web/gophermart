package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/divanov-web/gophermart/internal/middleware"
	"github.com/divanov-web/gophermart/internal/service"
	"go.uber.org/zap"
)

type OrderHandler struct {
	service *service.OrderService
	logger  *zap.SugaredLogger
}

func NewOrderHandler(service *service.OrderService, logger *zap.SugaredLogger) *OrderHandler {
	return &OrderHandler{
		service: service,
		logger:  logger,
	}
}

// Upload обрабатывает загрузку номера заказа
func (h *OrderHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "empty request", http.StatusBadRequest)
		return
	}
	orderNumber := strings.TrimSpace(string(body))

	err = h.service.UploadOrder(r.Context(), userID, orderNumber)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrOrderExists):
			w.WriteHeader(http.StatusOK)
		case errors.Is(err, service.ErrInvalidOrderNumber):
			http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
		case errors.Is(err, service.ErrOrderOwnedByAnotherUser):
			http.Error(w, "order already uploaded by another user", http.StatusConflict)
		default:
			h.logger.Errorw("upload order failed", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	} else {
		w.WriteHeader(http.StatusAccepted)
	}
}

func (h *OrderHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.service.GetOrdersByUser(r.Context(), userID)
	if err != nil {
		h.logger.Errorw("failed to fetch user orders", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "serialization error", http.StatusInternalServerError)
	}
}
