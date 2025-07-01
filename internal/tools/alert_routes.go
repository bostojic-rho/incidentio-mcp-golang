package tools

import (
	"encoding/json"
	"fmt"

	"github.com/incident-io/incidentio-mcp-golang/internal/incidentio"
)

// ListAlertRoutesTool lists alert routes from incident.io
type ListAlertRoutesTool struct {
	client *incidentio.Client
}

func NewListAlertRoutesTool(client *incidentio.Client) *ListAlertRoutesTool {
	return &ListAlertRoutesTool{client: client}
}

func (t *ListAlertRoutesTool) Name() string {
	return "list_alert_routes"
}

func (t *ListAlertRoutesTool) Description() string {
	return "List alert routes from incident.io with optional pagination"
}

func (t *ListAlertRoutesTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"page_size": map[string]interface{}{
				"type":        "integer",
				"description": "Number of results per page",
				"minimum":     1,
				"maximum":     250,
			},
			"after": map[string]interface{}{
				"type":        "string",
				"description": "Pagination cursor for next page",
			},
		},
		"additionalProperties": false,
	}
}

func (t *ListAlertRoutesTool) Execute(args map[string]interface{}) (string, error) {
	params := &incidentio.ListAlertRoutesParams{}

	if pageSize, ok := args["page_size"].(float64); ok {
		params.PageSize = int(pageSize)
	}
	if after, ok := args["after"].(string); ok {
		params.After = after
	}

	result, err := t.client.ListAlertRoutes(params)
	if err != nil {
		return "", fmt.Errorf("failed to list alert routes: %w", err)
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// GetAlertRouteTool gets details of a specific alert route
type GetAlertRouteTool struct {
	client *incidentio.Client
}

func NewGetAlertRouteTool(client *incidentio.Client) *GetAlertRouteTool {
	return &GetAlertRouteTool{client: client}
}

func (t *GetAlertRouteTool) Name() string {
	return "get_alert_route"
}

func (t *GetAlertRouteTool) Description() string {
	return "Get details of a specific alert route by ID"
}

func (t *GetAlertRouteTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "The alert route ID",
				"minLength":   1,
			},
		},
		"required":             []string{"id"},
		"additionalProperties": false,
	}
}

func (t *GetAlertRouteTool) Execute(args map[string]interface{}) (string, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("alert route ID is required")
	}

	alertRoute, err := t.client.GetAlertRoute(id)
	if err != nil {
		return "", fmt.Errorf("failed to get alert route: %w", err)
	}

	output, err := json.MarshalIndent(alertRoute, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// CreateAlertRouteTool creates a new alert route
type CreateAlertRouteTool struct {
	client *incidentio.Client
}

func NewCreateAlertRouteTool(client *incidentio.Client) *CreateAlertRouteTool {
	return &CreateAlertRouteTool{client: client}
}

func (t *CreateAlertRouteTool) Name() string {
	return "create_alert_route"
}

func (t *CreateAlertRouteTool) Description() string {
	return "Create a new alert route in incident.io"
}

func (t *CreateAlertRouteTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the alert route",
				"minLength":   1,
			},
			"enabled": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the alert route should be enabled",
				"default":     true,
			},
			"conditions": map[string]interface{}{
				"type":        "array",
				"description": "Conditions for routing alerts",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"field": map[string]interface{}{
							"type":        "string",
							"description": "Field to match on",
						},
						"operation": map[string]interface{}{
							"type":        "string",
							"description": "Operation to perform (equals, contains, etc)",
						},
						"value": map[string]interface{}{
							"type":        "string",
							"description": "Value to match against",
						},
					},
					"required": []string{"field", "operation", "value"},
				},
			},
			"escalations": map[string]interface{}{
				"type":        "array",
				"description": "Escalation bindings",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":        "string",
							"description": "Escalation ID",
						},
						"level": map[string]interface{}{
							"type":        "integer",
							"description": "Escalation level",
						},
					},
					"required": []string{"id", "level"},
				},
			},
			"grouping_keys": map[string]interface{}{
				"type":        "array",
				"description": "Keys to group alerts by",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"template": map[string]interface{}{
				"type":        "object",
				"description": "Template for creating incidents from alerts",
			},
		},
		"required":             []string{"name", "conditions", "escalations"},
		"additionalProperties": false,
	}
}

func (t *CreateAlertRouteTool) Execute(args map[string]interface{}) (string, error) {
	req := &incidentio.CreateAlertRouteRequest{}

	name, ok := args["name"].(string)
	if !ok || name == "" {
		return "", fmt.Errorf("name is required")
	}
	req.Name = name

	if enabled, ok := args["enabled"].(bool); ok {
		req.Enabled = enabled
	} else {
		req.Enabled = true // default to enabled
	}

	// Parse conditions
	if conditions, ok := args["conditions"].([]interface{}); ok {
		for _, c := range conditions {
			if cond, ok := c.(map[string]interface{}); ok {
				condition := incidentio.AlertCondition{
					Field:     cond["field"].(string),
					Operation: cond["operation"].(string),
					Value:     cond["value"].(string),
				}
				req.Conditions = append(req.Conditions, condition)
			}
		}
	}

	// Parse escalations
	if escalations, ok := args["escalations"].([]interface{}); ok {
		for _, e := range escalations {
			if esc, ok := e.(map[string]interface{}); ok {
				escalation := incidentio.EscalationBinding{
					ID:    esc["id"].(string),
					Level: int(esc["level"].(float64)),
				}
				req.Escalations = append(req.Escalations, escalation)
			}
		}
	}

	// Parse grouping keys
	if groupingKeys, ok := args["grouping_keys"].([]interface{}); ok {
		for _, k := range groupingKeys {
			if key, ok := k.(string); ok {
				req.GroupingKeys = append(req.GroupingKeys, key)
			}
		}
	}

	// Parse template
	if template, ok := args["template"].(map[string]interface{}); ok {
		req.Template = template
	}

	alertRoute, err := t.client.CreateAlertRoute(req)
	if err != nil {
		return "", fmt.Errorf("failed to create alert route: %w", err)
	}

	output, err := json.MarshalIndent(alertRoute, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// UpdateAlertRouteTool updates an alert route
type UpdateAlertRouteTool struct {
	client *incidentio.Client
}

func NewUpdateAlertRouteTool(client *incidentio.Client) *UpdateAlertRouteTool {
	return &UpdateAlertRouteTool{client: client}
}

func (t *UpdateAlertRouteTool) Name() string {
	return "update_alert_route"
}

func (t *UpdateAlertRouteTool) Description() string {
	return "Update an alert route's configuration"
}

func (t *UpdateAlertRouteTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "The alert route ID to update",
				"minLength":   1,
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "New name for the alert route",
			},
			"enabled": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the alert route should be enabled",
			},
			"conditions": map[string]interface{}{
				"type":        "array",
				"description": "New conditions for routing alerts",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"field": map[string]interface{}{
							"type":        "string",
							"description": "Field to match on",
						},
						"operation": map[string]interface{}{
							"type":        "string",
							"description": "Operation to perform",
						},
						"value": map[string]interface{}{
							"type":        "string",
							"description": "Value to match against",
						},
					},
					"required": []string{"field", "operation", "value"},
				},
			},
			"escalations": map[string]interface{}{
				"type":        "array",
				"description": "New escalation bindings",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":        "string",
							"description": "Escalation ID",
						},
						"level": map[string]interface{}{
							"type":        "integer",
							"description": "Escalation level",
						},
					},
					"required": []string{"id", "level"},
				},
			},
			"grouping_keys": map[string]interface{}{
				"type":        "array",
				"description": "Keys to group alerts by",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"template": map[string]interface{}{
				"type":        "object",
				"description": "Template for creating incidents from alerts",
			},
		},
		"required":             []string{"id"},
		"additionalProperties": false,
	}
}

func (t *UpdateAlertRouteTool) Execute(args map[string]interface{}) (string, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("alert route ID is required")
	}

	req := &incidentio.UpdateAlertRouteRequest{}

	if name, ok := args["name"].(string); ok {
		req.Name = name
	}

	if enabled, ok := args["enabled"].(bool); ok {
		req.Enabled = &enabled
	}

	// Parse conditions
	if conditions, ok := args["conditions"].([]interface{}); ok {
		req.Conditions = []incidentio.AlertCondition{}
		for _, c := range conditions {
			if cond, ok := c.(map[string]interface{}); ok {
				condition := incidentio.AlertCondition{
					Field:     cond["field"].(string),
					Operation: cond["operation"].(string),
					Value:     cond["value"].(string),
				}
				req.Conditions = append(req.Conditions, condition)
			}
		}
	}

	// Parse escalations
	if escalations, ok := args["escalations"].([]interface{}); ok {
		req.Escalations = []incidentio.EscalationBinding{}
		for _, e := range escalations {
			if esc, ok := e.(map[string]interface{}); ok {
				escalation := incidentio.EscalationBinding{
					ID:    esc["id"].(string),
					Level: int(esc["level"].(float64)),
				}
				req.Escalations = append(req.Escalations, escalation)
			}
		}
	}

	// Parse grouping keys
	if groupingKeys, ok := args["grouping_keys"].([]interface{}); ok {
		req.GroupingKeys = []string{}
		for _, k := range groupingKeys {
			if key, ok := k.(string); ok {
				req.GroupingKeys = append(req.GroupingKeys, key)
			}
		}
	}

	// Parse template
	if template, ok := args["template"].(map[string]interface{}); ok {
		req.Template = template
	}

	alertRoute, err := t.client.UpdateAlertRoute(id, req)
	if err != nil {
		return "", fmt.Errorf("failed to update alert route: %w", err)
	}

	output, err := json.MarshalIndent(alertRoute, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}
