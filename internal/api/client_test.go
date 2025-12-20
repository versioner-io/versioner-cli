package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleAPIError_PreflightErrorsAlwaysFail(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"409 Conflict", 409},
		{"423 Locked", 423},
		{"428 Precondition Required", 428},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				FailOnAPIError: false,
			}

			apiError := &APIError{
				StatusCode: tt.statusCode,
				Detail:     "preflight check failed",
			}

			var result DeploymentResponse
			err := client.handleAPIError(apiError, &result)

			if err == nil {
				t.Errorf("Expected error for status %d, got nil", tt.statusCode)
			}

			if err != apiError {
				t.Errorf("Expected original APIError to be returned")
			}
		})
	}
}

func TestHandleAPIError_APIErrorsRespectFlag(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		failOnAPIError bool
		expectError    bool
	}{
		{"401 with fail=true", 401, true, true},
		{"401 with fail=false", 401, false, false},
		{"403 with fail=true", 403, true, true},
		{"403 with fail=false", 403, false, false},
		{"404 with fail=true", 404, true, true},
		{"404 with fail=false", 404, false, false},
		{"422 with fail=true", 422, true, true},
		{"422 with fail=false", 422, false, false},
		{"500 with fail=true", 500, true, true},
		{"500 with fail=false", 500, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				FailOnAPIError: tt.failOnAPIError,
			}

			apiError := &APIError{
				StatusCode: tt.statusCode,
				Detail:     "API error",
			}

			var result DeploymentResponse
			err := client.handleAPIError(apiError, &result)

			if tt.expectError && err == nil {
				t.Errorf("Expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if !tt.expectError {
				if result.Status != "not_recorded" {
					t.Errorf("Expected placeholder response with status='not_recorded', got: %s", result.Status)
				}
			}
		})
	}
}

func TestHandleAPIError_PlaceholderResponses(t *testing.T) {
	client := &Client{
		FailOnAPIError: false,
	}

	apiError := &APIError{
		StatusCode: 401,
		Detail:     "Unauthorized",
	}

	t.Run("DeploymentResponse", func(t *testing.T) {
		var result DeploymentResponse
		err := client.handleAPIError(apiError, &result)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.Status != "not_recorded" {
			t.Errorf("Expected status='not_recorded', got: %s", result.Status)
		}

		if result.ID != "" || result.ProductID != "" || result.VersionID != "" || result.EnvironmentID != "" {
			t.Errorf("Expected empty IDs in placeholder response")
		}
	})

	t.Run("BuildResponse", func(t *testing.T) {
		var result BuildResponse
		err := client.handleAPIError(apiError, &result)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.Status != "not_recorded" {
			t.Errorf("Expected status='not_recorded', got: %s", result.Status)
		}

		if result.ID != "" || result.ProductID != "" || result.VersionID != "" {
			t.Errorf("Expected empty IDs in placeholder response")
		}
	})
}

func TestCreateDeploymentEvent_WithFailOnAPIError(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		failOnAPIError bool
		expectError    bool
	}{
		{"401 with fail=true", 401, true, true},
		{"401 with fail=false", 401, false, false},
		{"409 with fail=false (preflight)", 409, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(`{"detail": "error message"}`))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-key", false, tt.failOnAPIError)

			event := &DeploymentEventCreate{
				ProductName:     "test-product",
				Version:         "1.0.0",
				EnvironmentName: "production",
				Status:          "started",
			}

			resp, err := client.CreateDeploymentEvent(event)

			if tt.expectError && err == nil {
				t.Errorf("Expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if !tt.expectError && resp.Status != "not_recorded" {
				t.Errorf("Expected placeholder response with status='not_recorded', got: %s", resp.Status)
			}
		})
	}
}

func TestCreateBuildEvent_WithFailOnAPIError(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		failOnAPIError bool
		expectError    bool
	}{
		{"401 with fail=true", 401, true, true},
		{"401 with fail=false", 401, false, false},
		{"422 with fail=true", 422, true, true},
		{"422 with fail=false", 422, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(`{"detail": "error message"}`))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-key", false, tt.failOnAPIError)

			event := &BuildEventCreate{
				ProductName: "test-product",
				Version:     "1.0.0",
				Status:      "completed",
			}

			resp, err := client.CreateBuildEvent(event)

			if tt.expectError && err == nil {
				t.Errorf("Expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if !tt.expectError && resp.Status != "not_recorded" {
				t.Errorf("Expected placeholder response with status='not_recorded', got: %s", resp.Status)
			}
		})
	}
}
