package handlers

import (
	"encoding/json"
	"errors"
	"github.com/divanov-web/gophermart/internal/middleware"
	"github.com/divanov-web/gophermart/internal/service"
	"go.uber.org/zap"
	"net/http"
)

type BalanceHandler struct {
	orderService *service.OrderService
	UserService  *service.UserService
	logger       *zap.SugaredLogger
}

func NewBalanceHandler(orderService *service.OrderService, userService *service.UserService, logger *zap.SugaredLogger) *BalanceHandler {
	return &BalanceHandler{
		orderService: orderService,
		UserService:  userService,
		logger:       logger,
	}
}

func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req service.WithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	err := h.orderService.Withdraw(r.Context(), userID, req)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, service.ErrInvalidWithdrawOrder):
		http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
	case errors.Is(err, service.ErrInsufficientFunds):
		http.Error(w, "not enough funds", http.StatusPaymentRequired) // 402
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	balance, err := h.UserService.GetUserBalance(r.Context(), userID)
	if err != nil {
		h.logger.Errorw("failed to get balance", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(balance); err != nil {
		h.logger.Errorw("failed to encode balance", "error", err)
		http.Error(w, "serialization error", http.StatusInternalServerError)
	}
}

func (h *BalanceHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.UserService.GetWithdrawals(r.Context(), userID)
	if err != nil {
		h.logger.Errorw("failed to get withdrawals", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent) // 204
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		h.logger.Errorw("failed to encode withdrawals", "error", err)
		http.Error(w, "serialization error", http.StatusInternalServerError)
	}
}
