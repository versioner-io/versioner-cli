package api

import "time"

// DeploymentEventCreate represents the request payload for creating a deployment event
type DeploymentEventCreate struct {
	ProductName         string                 `json:"product_name"`
	Version             string                 `json:"version"`
	EnvironmentName     string                 `json:"environment_name"`
	Status              string                 `json:"status"`
	SourceSystem        string                 `json:"source_system,omitempty"`
	BuildNumber         string                 `json:"build_number,omitempty"`
	SCMSha              string                 `json:"scm_sha,omitempty"`
	SCMRepository       string                 `json:"scm_repository,omitempty"`
	DeployURL           string                 `json:"deploy_url,omitempty"`
	InvokeID            string                 `json:"invoke_id,omitempty"`
	DeployedBy          string                 `json:"deployed_by,omitempty"`
	DeployedByEmail     string                 `json:"deployed_by_email,omitempty"`
	DeployedByName      string                 `json:"deployed_by_name,omitempty"`
	CompletedAt         *time.Time             `json:"completed_at,omitempty"`
	SkipPreflightChecks bool                   `json:"skip_preflight_checks,omitempty"`
	ExtraMetadata       map[string]interface{} `json:"extra_metadata,omitempty"`
}

// DeploymentResponse represents the response from creating a deployment event
type DeploymentResponse struct {
	ID            string     `json:"id"`
	ProductID     string     `json:"product_id"`
	VersionID     string     `json:"version_id"`
	EnvironmentID string     `json:"environment_id"`
	Status        string     `json:"status"`
	DeployedAt    *time.Time `json:"deployed_at,omitempty"`
}

// PreflightError represents a preflight check failure with detailed information
type PreflightError struct {
	StatusCode int
	Error      string                 `json:"error"`
	Message    string                 `json:"message"`
	Code       string                 `json:"code"`
	Details    map[string]interface{} `json:"details"`
	RetryAfter string                 `json:"retry_after,omitempty"`
}

// CreateDeploymentEvent sends a deployment event to the API
func (c *Client) CreateDeploymentEvent(event *DeploymentEventCreate) (*DeploymentResponse, error) {
	resp, err := c.doRequest("POST", "/deployment-events/", event)
	if err != nil {
		return nil, err
	}

	var result DeploymentResponse
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
