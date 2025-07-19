package accrual

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/divanov-web/gophermart/internal/model"
	"github.com/divanov-web/gophermart/internal/utils/ordergen"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	logger     *zap.SugaredLogger
}

func NewClient(baseURL string, logger *zap.SugaredLogger) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

type AccrualRequest struct {
	Order string             `json:"order"`
	Goods []model.OrderGoods `json:"goods"`
}

type AccrualResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

// SendOrder Отправка нового заказа на сервер accrual
func (c *Client) SendOrder(orderNumber string) error {
	reqBody := AccrualRequest{
		Order: orderNumber,
		Goods: ordergen.GenerateRandomGoods(),
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal accrual request: %w", err)
	}

	url := fmt.Sprintf("%s/api/orders", c.BaseURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("build accrual request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("accrual request failed: %w", err)
	}
	defer resp.Body.Close()

	c.logger.Infow(
		"Sending order",
		"Order", orderNumber,
		"status", resp.StatusCode,
	)

	switch resp.StatusCode {
	case http.StatusAccepted:
		return nil
	case http.StatusConflict:
		return nil
	default:
		return fmt.Errorf("unexpected status from accrual: %s", resp.Status)
	}
}

func (c *Client) GetOrderInfo(orderNumber string) (*AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.BaseURL, orderNumber)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build accrual status request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("accrual status request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil // заказ не зарегистрирован
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status from accrual: %s", resp.Status)
	}

	var result AccrualResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("decode accrual response: %w", err)
	}

	return &result, nil
}
