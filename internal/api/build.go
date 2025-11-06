package api

import "time"

// BuildEventCreate represents the request payload for creating a build event
type BuildEventCreate struct {
	ProductName   string                 `json:"product_name"`
	Version       string                 `json:"version"`
	Status        string                 `json:"status"`
	SourceSystem  string                 `json:"source_system,omitempty"`
	BuildNumber   string                 `json:"build_number,omitempty"`
	SCMSha        string                 `json:"scm_sha,omitempty"`
	SCMBranch     string                 `json:"scm_branch,omitempty"`
	SCMRepository string                 `json:"scm_repository,omitempty"`
	BuildURL      string                 `json:"build_url,omitempty"`
	InvokeID      string                 `json:"invoke_id,omitempty"`
	BuiltBy       string                 `json:"built_by,omitempty"`
	BuiltByEmail  string                 `json:"built_by_email,omitempty"`
	BuiltByName   string                 `json:"built_by_name,omitempty"`
	StartedAt     *time.Time             `json:"started_at,omitempty"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	ExtraMetadata map[string]interface{} `json:"extra_metadata,omitempty"`
}

// BuildResponse represents the response from creating a build event
type BuildResponse struct {
	ID          string     `json:"id"`
	ProductID   string     `json:"product_id"`
	VersionID   string     `json:"version_id"`
	Status      string     `json:"status"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// CreateBuildEvent sends a build event to the API
func (c *Client) CreateBuildEvent(event *BuildEventCreate) (*BuildResponse, error) {
	resp, err := c.doRequest("POST", "/build-events/", event)
	if err != nil {
		return nil, err
	}

	var result BuildResponse
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
