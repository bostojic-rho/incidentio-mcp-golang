package incidentio

import (
	"encoding/json"
	"fmt"
)

// Using Severity type from types.go

// ListSeveritiesResponse represents the response from listing severities
type ListSeveritiesResponse struct {
	Severities []Severity `json:"severities"`
}

// ListSeverities returns all severities
func (c *Client) ListSeverities() (*ListSeveritiesResponse, error) {
	respBody, err := c.doRequest("GET", "/severities", nil, nil)
	if err != nil {
		return nil, err
	}

	var response ListSeveritiesResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetSeverity retrieves a specific severity by ID
func (c *Client) GetSeverity(id string) (*Severity, error) {
	respBody, err := c.doRequest("GET", fmt.Sprintf("/severities/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Severity Severity `json:"severity"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.Severity, nil
}