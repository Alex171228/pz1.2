package authclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"pz1.2/shared/middleware"
)

type HTTPClient struct {
	httpClient *http.Client
	baseURL    string
}

type VerifyResponse struct {
	Valid   bool   `json:"valid"`
	Subject string `json:"subject,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NewHTTPClient(baseURL string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
	}
}

func (c *HTTPClient) Verify(ctx context.Context, token string) (*VerifyResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.httpClient.Timeout)
	defer cancel()

	requestID := middleware.GetRequestID(ctx)
	log.Printf("[%s] Calling Auth HTTP verify", requestID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/auth/verify", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	if requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[%s] Auth HTTP verify failed: %v", requestID, err)
		return nil, fmt.Errorf("auth service unavailable: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var verifyResp VerifyResponse
	if err := json.Unmarshal(body, &verifyResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		log.Printf("[%s] Auth HTTP verify: unauthorized", requestID)
		return &verifyResp, nil
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[%s] Auth HTTP verify: unexpected status %d", requestID, resp.StatusCode)
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	log.Printf("[%s] Auth HTTP verify: success, subject=%s", requestID, verifyResp.Subject)
	return &verifyResp, nil
}
