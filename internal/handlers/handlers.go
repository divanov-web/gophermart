package handlers

import (
	"encoding/json"
	"github.com/divanov-web/gophermart/internal/service"
	"net/http"
)

type Handler struct {
	Service *service.UserService
}

// DataRequest Входящие данные
type DataRequest struct {
	URL string `json:"url"`
}

// DataResponse Исходящие данные
type DataResponse struct {
	Result string `json:"result"`
}

func NewHandler(svc *service.UserService) *Handler {
	return &Handler{Service: svc}
}

func (h *Handler) UserRegister(w http.ResponseWriter, r *http.Request) {
	result := DataResponse{Result: "success"}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Ошибка сериализации ответа", http.StatusInternalServerError)
	}
}
