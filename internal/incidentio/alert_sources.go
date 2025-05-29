package incidentio

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

// AlertSource represents an alert source in incident.io
type AlertSource struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	ConfigType string    `json:"config_type"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ListAlertSourcesParams contains optional parameters for listing alert sources
type ListAlertSourcesParams struct {
	PageSize int
	After    string
}

// ListAlertSourcesResponse represents the response from listing alert sources
type ListAlertSourcesResponse struct {
	AlertSources []AlertSource `json:"alert_sources"`
	Pagination   struct {
		After    string `json:"after,omitempty"`
		PageSize int    `json:"page_size"`
	} `json:"pagination_info"`
}

// ListAlertSources returns all alert sources
func (c *Client) ListAlertSources(params *ListAlertSourcesParams) (*ListAlertSourcesResponse, error) {
	endpoint := "/alert_sources"

	v := url.Values{}
	if params != nil {
		if params.PageSize > 0 {
			v.Set("page_size", fmt.Sprintf("%d", params.PageSize))
		}
		if params.After != "" {
			v.Set("after", params.After)
		}
	}

	if len(v) > 0 {
		endpoint = endpoint + "?" + v.Encode()
	}

	respBody, err := c.doRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var result ListAlertSourcesResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}
