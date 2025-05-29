# incident.io MCP Server

[![CI](https://github.com/twentworth12/incidentio-mcp-golang/actions/workflows/ci.yml/badge.svg)](https://github.com/twentworth12/incidentio-mcp-golang/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/twentworth12/incidentio-mcp-golang)](https://goreportcard.com/report/github.com/twentworth12/incidentio-mcp-golang)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://go.dev/dl/)

A GoLang implementation of an MCP (Model Context Protocol) server for incident.io, providing tools to interact with the incident.io V2 API.

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/twentworth12/incidentio-mcp-golang.git
cd incidentio-mcp-golang

# Copy environment variables
cp .env.example .env
# Edit .env and add your incident.io API key

# Build and run
make build
./start-mcp-server.sh
```

## 📋 Features

- ✅ Complete incident.io V2 API coverage
- ✅ Workflow automation and management
- ✅ Alert routing and event handling
- ✅ Comprehensive test suite
- ✅ MCP protocol compliant
- ✅ Clean, modular architecture

## Project Structure

```
.
├── cmd/mcp-server/      # Main application entry point
├── internal/            # Private application code
│   ├── server/          # MCP server implementation
│   └── tools/           # Tool implementations
├── pkg/mcp/             # MCP protocol types and utilities
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
└── Makefile             # Build commands
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- incident.io API key (set as `INCIDENT_IO_API_KEY` environment variable)

### Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   make deps
   ```

### Building

```bash
make build
```

This will create a binary in the `bin/` directory.

### Running

Set your incident.io API key:
```bash
export INCIDENT_IO_API_KEY=your-api-key
```

Then run the server:
```bash
make run
```

Or after building:
```bash
./bin/mcp-server
```

### Testing

```bash
make test
```

## Adding New Tools

1. Create a new file in `internal/tools/` implementing the `Tool` interface
2. Register the tool in `server.registerTools()` method in `internal/server/server.go`

Example tool implementation:
```go
type MyTool struct{}

func (t *MyTool) Name() string {
    return "my_tool"
}

func (t *MyTool) Description() string {
    return "Description of what the tool does"
}

func (t *MyTool) InputSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            // Define your parameters here
        },
        "required": []string{/* required parameters */},
    }
}

func (t *MyTool) Execute(args map[string]interface{}) (string, error) {
    // Tool implementation
    return "result", nil
}
```

## Available Tools

### Incident Management
- `list_incidents` - List incidents with optional filters (status, severity)
- `get_incident` - Get details of a specific incident by ID
- `create_incident` - Create a new incident
- `update_incident` - Update an existing incident
- `close_incident` - Close an incident with proper workflow handling
- `list_incident_statuses` - List available incident statuses
- `list_incident_types` - List available incident types

### Incident Updates
- `list_incident_updates` - List incident status updates/messages
- `get_incident_update` - Get a specific incident update
- `create_incident_update` - Post a new status update to an incident
- `delete_incident_update` - Delete an incident update

### Severity Management
- `list_severities` - List available severity levels
- `get_severity` - Get details of a specific severity

### Alert Management
- `list_alerts` - List alerts with optional filters
- `get_alert` - Get details of a specific alert by ID
- `list_alerts_for_incident` - List alerts associated with a specific incident
- `list_alert_sources` - List available alert sources
- `create_alert_event` - Create an alert event

### Alert Routing
- `list_alert_routes` - List alert routes with optional pagination
- `get_alert_route` - Get details of a specific alert route
- `create_alert_route` - Create a new alert route with conditions and escalations
- `update_alert_route` - Update an alert route's configuration

### Workflow Management
- `list_workflows` - List workflows with optional pagination
- `get_workflow` - Get details of a specific workflow
- `update_workflow` - Update a workflow's configuration

### Action Management
- `list_actions` - List actions with optional filters (incident_id, status)
- `get_action` - Get details of a specific action by ID

### Roles and Users
- `list_available_incident_roles` - List available incident roles
- `list_users` - List users in the organization
- `assign_incident_role` - Assign a role to a user for an incident

### Testing
- `example_tool` - A simple echo tool for testing

## MCP Protocol

This server implements the Model Context Protocol (MCP) for communication with AI assistants. The server:
- Communicates via JSON-RPC over stdin/stdout
- Supports tool registration and execution
- Follows the MCP 2024-11-05 protocol version
- Integrates with incident.io V2 API endpoints

## 🤖 Using with Claude

Add to your Claude configuration:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "incidentio": {
      "command": "/path/to/incidentio-mcp-golang/bin/mcp-server",
      "env": {
        "INCIDENT_IO_API_KEY": "your-api-key"
      }
    }
  }
}
```

After updating the configuration, restart Claude to load the incident.io tools.

## Environment Variables

- `INCIDENT_IO_API_KEY` (required) - Your incident.io API key
- `INCIDENT_IO_BASE_URL` (optional) - Override the API base URL (defaults to https://api.incident.io/v2)

## 🧪 Testing

```bash
# Run unit tests
make test-unit

# Run integration tests (requires API key)
make test-integration

# Run all tests
make test
```

See [TESTING.md](TESTING.md) for detailed testing documentation.

## 🔧 Troubleshooting

### Common Issues

- **404 errors**: Ensure incident IDs are valid and exist in your instance
- **Authentication errors**: Verify your API key is correct and has proper permissions
- **Parameter errors**: All incident-related tools use `incident_id` as the parameter name
- **Status transition errors**: Some incident status changes require specific workflows (e.g., must be in "Monitoring" before "Closed")
- **Page size errors**: Different endpoints have different limits (incidents: 250, alerts: 50)

### Debug Mode

To enable debug logging, use the wrapper script:

```bash
# Create wrapper script
cat > mcp-debug-wrapper.sh << 'EOF'
#!/bin/bash
LOG_FILE="/tmp/mcp-incidentio-debug-$(date +%Y%m%d).log"
exec /path/to/bin/mcp-server 2>>"$LOG_FILE"
EOF

chmod +x mcp-debug-wrapper.sh
```

Then use the wrapper in your Claude configuration to capture debug logs.

## 🤝 Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with the [Model Context Protocol](https://modelcontextprotocol.io/) specification
- Powered by [incident.io](https://incident.io/) API
- Created with assistance from Claude