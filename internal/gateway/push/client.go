package push

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// RequestBody — тело запроса к Push Service.
type RequestBody struct {
	Kind        string            `json:"kind"`
	RecipientID uuid.UUID         `json:"recipient_id"`
	Data        map[string]string `json:"data"`
}

// Client — HTTP-клиент для отправки пушей.
type Client struct {
	baseURL string
	client  *http.Client
}

// NewClient создает клиент. Таймаут задается снаружи (из конфига).
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) RequestPosition(ctx context.Context, driverID uuid.UUID, tripID uuid.UUID) error {
	body := RequestBody{
		Kind:        "REQUEST_POSITION",
		RecipientID: driverID,
		Data: map[string]string{
			"trip_id": tripID.String(),
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/push", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("push service returned status %d", resp.StatusCode)
	}

	return nil
}