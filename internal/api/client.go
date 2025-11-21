package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/versioner-io/versioner-cli/internal/version"
)

// Client represents the Versioner API client
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	UserAgent  string
	Debug      bool
}

// NewClient creates a new API client
func NewClient(baseURL, apiKey string, debug bool) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		UserAgent: version.GetUserAgent(),
		Debug:     debug,
	}
}

// doRequest performs an HTTP request with retry logic
func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var lastErr error

	// Retry logic: 3 attempts with exponential backoff
	backoffDurations := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

	for attempt := 0; attempt <= 3; attempt++ {
		if attempt > 0 {
			// Wait before retry
			time.Sleep(backoffDurations[attempt-1])
		}

		resp, err := c.performRequest(method, path, body)
		if err != nil {
			lastErr = err
			// Retry on network errors
			continue
		}

		// Success (2xx)
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		// Don't retry on 4xx errors (except 429 Too Many Requests)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != 429 {
			return resp, nil
		}

		// Retry on 5xx errors and 429
		lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
		resp.Body.Close()
	}

	return nil, fmt.Errorf("request failed after 3 retries: %w", lastErr)
}

// performRequest performs a single HTTP request
func (c *Client) performRequest(method, path string, body interface{}) (*http.Response, error) {
	url := c.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)

		if c.Debug {
			fmt.Printf("→ Request body: %s\n", string(jsonBody))
		}
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("User-Agent", c.UserAgent)

	if c.Debug {
		fmt.Printf("→ %s %s\n", method, url)
		fmt.Printf("→ Headers: %v\n", req.Header)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if c.Debug {
		fmt.Printf("← Status: %d\n", resp.StatusCode)
	}

	return resp, nil
}

// handleResponse processes the API response
func handleResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Success
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if result != nil {
			if err := json.Unmarshal(body, result); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
		}
		return nil
	}

	// Error response
	var errorResponse struct {
		Detail interface{} `json:"detail"`
	}
	if err := json.Unmarshal(body, &errorResponse); err != nil {
		// Fallback if error response doesn't match expected format
		return &APIError{
			StatusCode: resp.StatusCode,
			Detail:     string(body),
		}
	}

	apiError := &APIError{
		StatusCode: resp.StatusCode,
		Detail:     errorResponse.Detail,
	}

	return apiError
}

// APIError represents an error response from the API
type APIError struct {
	StatusCode int         `json:"-"`
	Detail     interface{} `json:"detail"`
}

func (e *APIError) Error() string {
	switch detail := e.Detail.(type) {
	case string:
		return detail
	case []interface{}:
		// Validation errors
		if len(detail) > 0 {
			return fmt.Sprintf("validation error: %v", detail)
		}
		return "validation error"
	default:
		return fmt.Sprintf("API error: %v", detail)
	}
}

// IsPreflightError checks if this is a preflight check failure (409, 423, 428)
func (e *APIError) IsPreflightError() bool {
	return e.StatusCode == 409 || e.StatusCode == 423 || e.StatusCode == 428
}

// GetPreflightDetails extracts structured preflight error details
func (e *APIError) GetPreflightDetails() (errorType, message, code, retryAfter string, details map[string]interface{}, ok bool) {
	detailMap, ok := e.Detail.(map[string]interface{})
	if !ok {
		return
	}

	if errType, exists := detailMap["error"].(string); exists {
		errorType = errType
	}
	if msg, exists := detailMap["message"].(string); exists {
		message = msg
	}
	if c, exists := detailMap["code"].(string); exists {
		code = c
	}
	if retry, exists := detailMap["retry_after"].(string); exists {
		retryAfter = retry
	}
	if det, exists := detailMap["details"].(map[string]interface{}); exists {
		details = det
	}

	ok = true
	return
}
