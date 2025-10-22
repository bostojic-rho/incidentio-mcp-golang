package handlers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/incident-io/incidentio-mcp-golang/internal/client"
)

// ListIncidentsTool lists incidents from incident.io
type ListIncidentsTool struct {
	apiClient *client.Client
}

func NewListIncidentsTool(c *client.Client) *ListIncidentsTool {
	return &ListIncidentsTool{apiClient: c}
}

func (t *ListIncidentsTool) Name() string {
	return "list_incidents"
}

func (t *ListIncidentsTool) Description() string {
	return "List incidents with server-side filtering. Returns paginated results - you MUST fetch ALL pages.\n\n" +
		"CRITICAL PAGINATION RULE:\n" +
		"If user asks to 'list all' or 'show all' incidents, you MUST paginate through ALL pages automatically.\n" +
		"- Start: page_size=100 (optimal for efficiency)\n" +
		"- If response has_more_results=true: IMMEDIATELY call list_incidents again with after cursor\n" +
		"- Repeat until has_more_results=false\n" +
		"- Only then provide the complete results to user\n" +
		"DO NOT show partial results and say 'there might be more' - fetch everything first!\n\n" +
		"INCIDENT REFERENCE RESOLUTION:\n" +
		"If user mentions specific incident references (INC-1691), use get_incident({\"incident_id\": \"1691\"}) first to get details.\n" +
		"For follow-ups/updates on specific incidents, you'll need the actual incident ID from get_incident response.\n\n" +
		"TEAM/CUSTOM FIELD FILTERING:\n" +
		"1. search_custom_fields({\"query\": \"team\"}) → get field ID and options\n" +
		"2. Find option where value=\"Engineering\" → get option.id\n" +
		"3. list_incidents with custom_field_id + custom_field_value (use option ID!)\n\n" +
		"EXAMPLES:\n" +
		"User: 'show all Engineering team incidents from past week'\n" +
		"→ search_custom_fields({\"query\": \"team\"})\n" +
		"→ list_incidents({\"custom_field_id\": \"cf_ABC\", \"custom_field_value\": \"opt_XYZ\", \"created_at_gte\": \"2025-10-08\", \"page_size\": 100})\n" +
		"→ IF has_more_results=true: list_incidents({same filters, \"after\": cursor}) and repeat\n" +
		"→ Combine all pages, then show user complete list\n\n" +
		"Date format: \"2025-10-15\". IMPORTANT: Use current year when calculating dates!"
}

func (t *ListIncidentsTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"page_size": map[string]interface{}{
				"type":        "integer",
				"description": "Number of results per page. Use 50-100 for efficiency. To get more results, use pagination with 'after' parameter.",
				"default":     100,
				"minimum":     1,
				"maximum":     100,
			},
			"after": map[string]interface{}{
				"type":        "string",
				"description": "Pagination cursor from previous response. Get this from pagination_meta.after in the previous response.",
			},
			"status": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Filter by status. Values: triage, active, investigating, monitoring, resolved, closed. Example: ['active', 'triage']",
			},
			"severity_one_of": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Filter by exact severity IDs. Use list_severities to get IDs. Example: ['01ABC123']",
			},
			"severity_gte": map[string]interface{}{
				"type":        "string",
				"description": "Filter by severity rank >= this ID. Returns this severity and all more severe. Example: 'sev_major_id' returns Major, Critical.",
			},
			"severity_lte": map[string]interface{}{
				"type":        "string",
				"description": "Filter by severity rank <= this ID. Returns this severity and all less severe. Example: 'sev_major_id' returns Major, Minor, Low.",
			},
			"created_at_gte": map[string]interface{}{
				"type":        "string",
				"description": "Filter incidents created on or after this date. Format: '2025-10-15' or '2025-10-15T10:30:00Z'. Use current year (2025).",
			},
			"created_at_lte": map[string]interface{}{
				"type":        "string",
				"description": "Filter incidents created on or before this date. Format: '2025-10-15' or '2025-10-15T23:59:59Z'. Use current year (2025).",
			},
			"updated_at_gte": map[string]interface{}{
				"type":        "string",
				"description": "Filter incidents updated on or after this date. Format: '2025-10-15' or '2025-10-15T10:30:00Z'. Use current year (2025).",
			},
			"updated_at_lte": map[string]interface{}{
				"type":        "string",
				"description": "Filter incidents updated on or before this date. Format: '2025-10-15' or '2025-10-15T23:59:59Z'. Use current year (2025).",
			},
			"custom_field_id": map[string]interface{}{
				"type":        "string",
				"description": "Custom field ID to filter by. Must use with custom_field_value. Get ID from search_custom_fields.",
			},
			"custom_field_value": map[string]interface{}{
				"type":        "string",
				"description": "Custom field OPTION ID to match. For select fields, this must be the option's ID (e.g., '01JQ7...'), not the label. Get from the options array of search_custom_fields response.",
			},
		},
	}
}

func (t *ListIncidentsTool) Execute(args map[string]interface{}) (string, error) {
	opts := &client.ListIncidentsOptions{
		PageSize: 100, // Default to 100 for better efficiency
	}

	if pageSize, ok := args["page_size"].(float64); ok {
		opts.PageSize = int(pageSize)
	}

	if after, ok := args["after"].(string); ok && after != "" {
		opts.After = after
	}

	if statuses, ok := args["status"].([]interface{}); ok {
		for _, s := range statuses {
			if str, ok := s.(string); ok {
				opts.Status = append(opts.Status, str)
			}
		}
	}

	if severities, ok := args["severity_one_of"].([]interface{}); ok {
		for _, s := range severities {
			if str, ok := s.(string); ok {
				opts.SeverityOneOf = append(opts.SeverityOneOf, str)
			}
		}
	}

	if severityGte, ok := args["severity_gte"].(string); ok && severityGte != "" {
		opts.SeverityGte = severityGte
	}

	if severityLte, ok := args["severity_lte"].(string); ok && severityLte != "" {
		opts.SeverityLte = severityLte
	}

	if createdAtGte, ok := args["created_at_gte"].(string); ok && createdAtGte != "" {
		opts.CreatedAtGte = createdAtGte
	}

	if createdAtLte, ok := args["created_at_lte"].(string); ok && createdAtLte != "" {
		opts.CreatedAtLte = createdAtLte
	}

	if updatedAtGte, ok := args["updated_at_gte"].(string); ok && updatedAtGte != "" {
		opts.UpdatedAtGte = updatedAtGte
	}

	if updatedAtLte, ok := args["updated_at_lte"].(string); ok && updatedAtLte != "" {
		opts.UpdatedAtLte = updatedAtLte
	}

	// Handle custom field filtering - API format: custom_field[ID][one_of]=option_id
	if customFieldID, ok := args["custom_field_id"].(string); ok && customFieldID != "" {
		if customFieldValue, ok := args["custom_field_value"].(string); ok && customFieldValue != "" {
			if opts.CustomFieldOneOf == nil {
				opts.CustomFieldOneOf = make(map[string]string)
			}
			opts.CustomFieldOneOf[customFieldID] = customFieldValue
		}
	}

	resp, err := t.apiClient.ListIncidents(opts)
	if err != nil {
		return "", err
	}

	// Create response with prominent pagination info
	response := map[string]interface{}{
		"incidents":       resp.Incidents,
		"pagination_meta": resp.PaginationMeta,
		"count":           len(resp.Incidents),
	}

	// Add prominent pagination status
	// Use total_record_count to determine if there are more results
	// The "after" cursor is only needed for the next API call, not for determining if more results exist
	recordsFetched := len(resp.Incidents)
	totalRecords := resp.PaginationMeta.TotalRecordCount
	hasMore := recordsFetched < totalRecords

	if hasMore {
		response["has_more_results"] = true
		response["pagination_progress"] = map[string]interface{}{
			"records_fetched":  recordsFetched,
			"total_records":    totalRecords,
			"remaining":        totalRecords - recordsFetched,
			"progress_percent": fmt.Sprintf("%.1f%%", float64(recordsFetched)/float64(totalRecords)*100),
		}
		response["FETCH_NEXT_PAGE"] = map[string]interface{}{
			"action":  "REQUIRED - You must call list_incidents again to get remaining results",
			"after":   resp.PaginationMeta.After,
			"message": fmt.Sprintf("Fetched %d of %d incidents (%.1f%%). Call list_incidents again with after='%s' plus same filters. Repeat until has_more_results=false.", recordsFetched, totalRecords, float64(recordsFetched)/float64(totalRecords)*100, resp.PaginationMeta.After),
		}
	} else {
		response["has_more_results"] = false
		response["pagination_progress"] = map[string]interface{}{
			"records_fetched":  recordsFetched,
			"total_records":    totalRecords,
			"remaining":        0,
			"progress_percent": "100.0%",
		}
		response["pagination_status"] = fmt.Sprintf("COMPLETE - All %d incidents fetched", totalRecords)
	}

	return FormatJSONResponse(response)
}

// GetIncidentTool retrieves a specific incident
type GetIncidentTool struct {
	apiClient *client.Client
}

func NewGetIncidentTool(c *client.Client) *GetIncidentTool {
	return &GetIncidentTool{apiClient: c}
}

func (t *GetIncidentTool) Name() string {
	return "get_incident"
}

func (t *GetIncidentTool) Description() string {
	return "Get details of a specific incident by ID or reference.\n\n" +
		"ACCEPTS BOTH:\n" +
		"- Full incident ID: '01K3VHM0T0ZTMG9JPJ9GESB7XX'\n" +
		"- Short reference: '1691' (just the number from INC-1691)\n\n" +
		"USE THIS TOOL TO:\n" +
		"- Resolve incident references (INC-1691 → use '1691') to get full incident details\n" +
		"- Get the FULL incident ID needed for other API calls (follow-ups, updates, etc.)\n" +
		"- The response contains the full incident ID that other endpoints require\n\n" +
		"EXAMPLES:\n" +
		"- get_incident({\"incident_id\": \"1691\"}) - Get incident by reference number\n" +
		"- get_incident({\"incident_id\": \"01K3VHM0T0ZTMG9JPJ9GESB7XX\"}) - Get incident by full ID"
}

func (t *GetIncidentTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"incident_id": map[string]interface{}{
				"type":        "string",
				"description": "The incident ID or reference number. Accepts both full IDs (01K3VHM0T0ZTMG9JPJ9GESB7XX) and reference numbers (1691 from INC-1691).",
			},
		},
		"required":             []interface{}{"incident_id"},
		"additionalProperties": false,
	}
}

func (t *GetIncidentTool) Execute(args map[string]interface{}) (string, error) {
	id, ok := args["incident_id"].(string)
	if !ok || id == "" {
		argDetails := make(map[string]interface{})
		for key, value := range args {
			argDetails[key] = value
		}
		return "", fmt.Errorf("incident_id parameter is required and must be a non-empty string. Received parameters: %+v", argDetails)
	}

	incident, err := t.apiClient.GetIncident(id)
	if err != nil {
		return "", err
	}

	result, err := json.MarshalIndent(incident, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}

// CreateIncidentTool creates a new incident
type CreateIncidentTool struct {
	apiClient *client.Client
}

func NewCreateIncidentTool(c *client.Client) *CreateIncidentTool {
	return &CreateIncidentTool{apiClient: c}
}

func (t *CreateIncidentTool) Name() string {
	return "create_incident"
}

func (t *CreateIncidentTool) Description() string {
	return "Create a new incident in incident.io"
}

func (t *CreateIncidentTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "The incident name/title",
			},
			"summary": map[string]interface{}{
				"type":        "string",
				"description": "A summary of the incident",
			},
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Initial status (triage, active, resolved, closed)",
				"default":     "triage",
			},
			"severity_id": map[string]interface{}{
				"type":        "string",
				"description": "The severity ID",
			},
			"incident_type_id": map[string]interface{}{
				"type":        "string",
				"description": "The incident type ID",
			},
			"incident_status_id": map[string]interface{}{
				"type":        "string",
				"description": "The incident status ID",
			},
			"mode": map[string]interface{}{
				"type":        "string",
				"description": "The incident mode (standard, retrospective, tutorial)",
				"enum":        []string{"standard", "retrospective", "tutorial"},
				"default":     "standard",
			},
			"visibility": map[string]interface{}{
				"type":        "string",
				"description": "The incident visibility (public, private)",
				"enum":        []string{"public", "private"},
				"default":     "public",
			},
			"slack_channel_name_override": map[string]interface{}{
				"type":        "string",
				"description": "Override the auto-generated Slack channel name",
			},
		},
		"required":             []interface{}{"name"},
		"additionalProperties": false,
	}
}

func (t *CreateIncidentTool) Execute(args map[string]interface{}) (string, error) {
	name, ok := args["name"].(string)
	if !ok {
		return "", fmt.Errorf("name parameter is required")
	}

	// Generate idempotency key using timestamp and name
	idempotencyKey := fmt.Sprintf("mcp-%d-%s", time.Now().UnixNano(), name)

	req := &client.CreateIncidentRequest{
		IdempotencyKey: idempotencyKey,
		Name:           name,
		Mode:           "standard", // Default to standard mode
		Visibility:     "public",   // Default to public visibility
	}

	if summary, ok := args["summary"].(string); ok {
		req.Summary = summary
	}
	if statusID, ok := args["incident_status_id"].(string); ok {
		req.IncidentStatusID = statusID
	}
	if severityID, ok := args["severity_id"].(string); ok {
		req.SeverityID = severityID
	}
	if typeID, ok := args["incident_type_id"].(string); ok {
		req.IncidentTypeID = typeID
	}
	if mode, ok := args["mode"].(string); ok {
		req.Mode = mode
	}
	if visibility, ok := args["visibility"].(string); ok {
		req.Visibility = visibility
	}
	if slackOverride, ok := args["slack_channel_name_override"].(string); ok {
		req.SlackChannelNameOverride = slackOverride
	}

	// Check if critical fields are missing and provide helpful suggestions
	var suggestions []string

	if req.SeverityID == "" {
		suggestions = append(suggestions, "severity_id is not set. Use list_severities to see available options.")
	}

	if req.IncidentTypeID == "" {
		suggestions = append(suggestions, "incident_type_id is not set. Use list_incident_types to see available options.")
	}

	if req.IncidentStatusID == "" {
		suggestions = append(suggestions, "incident_status_id is not set. Use list_incident_statuses to see available options.")
	}

	incident, err := t.apiClient.CreateIncident(req)
	if err != nil {
		// If the error is related to missing required fields, provide more helpful error message
		errMsg := err.Error()
		if len(suggestions) > 0 && (strings.Contains(errMsg, "severity") || strings.Contains(errMsg, "incident_type") || strings.Contains(errMsg, "incident_status")) {
			return "", fmt.Errorf("%s\n\nSuggestions:\n%s", errMsg, strings.Join(suggestions, "\n"))
		}
		return "", err
	}

	// Include suggestions in successful response if fields were missing
	result, err := json.MarshalIndent(incident, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	if len(suggestions) > 0 {
		return fmt.Sprintf("%s\n\nNote: Incident created with defaults. %s", result, strings.Join(suggestions, " ")), nil
	}

	return string(result), nil
}

// UpdateIncidentTool updates an existing incident
type UpdateIncidentTool struct {
	apiClient *client.Client
}

func NewUpdateIncidentTool(c *client.Client) *UpdateIncidentTool {
	return &UpdateIncidentTool{apiClient: c}
}

func (t *UpdateIncidentTool) Name() string {
	return "update_incident"
}

func (t *UpdateIncidentTool) Description() string {
	return "Update an existing incident"
}

func (t *UpdateIncidentTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"incident_id": map[string]interface{}{
				"type":        "string",
				"description": "The incident ID to update",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Update the incident name",
			},
			"summary": map[string]interface{}{
				"type":        "string",
				"description": "Update the incident summary",
			},
			"incident_status_id": map[string]interface{}{
				"type":        "string",
				"description": "Update the incident status ID",
			},
			"severity_id": map[string]interface{}{
				"type":        "string",
				"description": "Update the severity ID",
			},
		},
		"required":             []interface{}{"incident_id"},
		"additionalProperties": false,
	}
}

func (t *UpdateIncidentTool) Execute(args map[string]interface{}) (string, error) {

	id, ok := args["incident_id"].(string)
	if !ok || id == "" {
		argDetails := make(map[string]interface{})
		for key, value := range args {
			argDetails[key] = value
		}
		return "", fmt.Errorf("incident_id parameter is required and must be a non-empty string. Received parameters: %+v", argDetails)
	}

	req := &client.UpdateIncidentRequest{}
	hasUpdate := false

	if name, ok := args["name"].(string); ok {
		req.Name = name
		hasUpdate = true
	}
	if summary, ok := args["summary"].(string); ok {
		req.Summary = summary
		hasUpdate = true
	}
	if statusID, ok := args["incident_status_id"].(string); ok {
		req.IncidentStatusID = statusID
		hasUpdate = true
	}
	if severityID, ok := args["severity_id"].(string); ok {
		req.SeverityID = severityID
		hasUpdate = true
	}

	if !hasUpdate {
		return "", fmt.Errorf("at least one field to update must be provided")
	}

	incident, err := t.apiClient.UpdateIncident(id, req)
	if err != nil {
		return "", err
	}

	result, err := json.MarshalIndent(incident, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(result), nil
}
