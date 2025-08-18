# JiraCLI Library Examples

This directory contains example applications demonstrating how to use the jira-cli library in your Go projects.

## Examples

### 1. Basic Usage (`basic-usage/`)

A simple command-line tool that demonstrates core Jira operations:

- List issues
- Create issues
- View issue details
- Update issues
- Add comments
- Transition issues

### 2. Advanced CLI (`advanced-cli/`)

A full-featured interactive CLI application with:

- Configuration management
- Interactive mode
- Bulk operations
- Error handling
- Project management
- Sprint operations (placeholder)

## Prerequisites

- Go 1.21 or higher
- Jira account with API access
- API token (for Cloud) or password (for Server/Data Center)

## Getting API Token

### Jira Cloud

1. Log in to [Atlassian Account Settings](https://id.atlassian.com/manage-profile/security/api-tokens)
2. Click "Create API token"
3. Give it a name and click "Create"
4. Copy the token immediately (you won't be able to see it again)

### Jira Server/Data Center

Use your regular Jira password or Personal Access Token if configured.

## Running the Examples

### Basic Usage Example

```bash
cd examples/basic-usage

# Using command-line flags
go run main.go \
  -server=https://your-domain.atlassian.net \
  -login=your-email@example.com \
  -token=your-api-token \
  -project=PROJ \
  -action=list

# Using environment variables
export JIRA_SERVER=https://your-domain.atlassian.net
export JIRA_LOGIN=your-email@example.com
export JIRA_API_TOKEN=your-api-token
export JIRA_PROJECT=PROJ

go run main.go -action=list
```

Available actions:

- `list` - List recent issues (default)
- `create` - Create a new issue
- `view` - View an issue (requires `-issue` flag)
- `update` - Update an issue (requires `-issue` flag)
- `comment` - Add comment to an issue (requires `-issue` flag)
- `transition` - Transition an issue (requires `-issue` flag)

Example commands:

```bash
# List issues
go run main.go -action=list

# View a specific issue
go run main.go -action=view -issue=PROJ-123

# Create a new issue
go run main.go -action=create

# Add a comment
go run main.go -action=comment -issue=PROJ-123

# Transition an issue
go run main.go -action=transition -issue=PROJ-123
```

### Advanced CLI Example

```bash
cd examples/advanced-cli

# Run with configuration file
go run main.go -config=jira-config.json

# Run with environment variables
export JIRA_SERVER=https://your-domain.atlassian.net
export JIRA_LOGIN=your-email@example.com
export JIRA_API_TOKEN=your-api-token
export JIRA_PROJECT=PROJ
go run main.go

# Run with debug mode
go run main.go -debug
```

The advanced CLI provides an interactive shell with commands:

```
> help                          # Show available commands
> search project = PROJ         # Search with JQL
> view PROJ-123                 # View issue details
> create                        # Create issue interactively
> update PROJ-123               # Update an issue
> comment PROJ-123              # Add a comment
> transition PROJ-123           # Change issue status
> assign PROJ-123 john@example  # Assign to user
> projects                      # List all projects
> bulk assign PROJ-1 PROJ-2     # Bulk operations
> quit                          # Exit
```

## Configuration File Format

For the advanced CLI, you can use a JSON configuration file:

```json
{
  "server": "https://your-domain.atlassian.net",
  "login": "your-email@example.com",
  "api_token": "your-api-token",
  "default_project": "PROJ",
  "installation_type": "Cloud"
}
```

Save this as `jira-config.json` and use with `-config` flag.

## Building Standalone Binaries

```bash
# Build basic example
cd examples/basic-usage
go build -o jira-basic

# Build advanced CLI
cd examples/advanced-cli
go build -o jira-cli

# Run the binary
./jira-cli -config=jira-config.json
```

## Using as a Library in Your Project

To use the jira-cli library in your own project:

```bash
go get github.com/eliziario/jira-lib/lib
```

Then in your code:

```go
package main

import (
    "log"
    "github.com/eliziario/jira-lib/lib"
)

func main() {
    client, err := lib.NewClient(lib.ClientConfig{
        Server:   "https://your-domain.atlassian.net",
        Login:    "your-email@example.com",
        APIToken: "your-api-token",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Use the client
    issue, err := client.GetIssue("PROJ-123")
    if err != nil {
        log.Fatal(err)
    }
  
    // ... do something with the issue
}
```

## Error Handling

The library returns typed errors that you can handle appropriately:

```go
import "github.com/eliziario/jira-lib/pkg/jira"

issue, err := client.GetIssue("PROJ-999")
if err != nil {
    switch e := err.(type) {
    case *jira.ErrUnexpectedResponse:
        // Handle HTTP errors
        log.Printf("HTTP %d: %s", e.StatusCode, e.Body)
    default:
        // Handle other errors
        log.Printf("Error: %v", err)
    }
}
```

## Security Notes

1. **Never commit API tokens**: Use environment variables or secure configuration management
2. **Secure config files**: If saving tokens to files, ensure proper file permissions (0600)
3. **Use HTTPS**: Always use HTTPS URLs for Jira servers
4. **Rotate tokens**: Regularly rotate your API tokens
5. **Minimal permissions**: Create tokens with only necessary permissions

## Troubleshooting

### Connection Issues

- Verify your server URL (should include `https://`)
- Check your API token is valid
- For on-premise, check if you need to use VPN
- Try enabling debug mode with `-debug` flag

### Authentication Errors

- For Cloud: Ensure you're using an API token, not your password
- For Server: Try using your password or Personal Access Token
- Check your email/username is correct

### Permission Errors

- Ensure your account has necessary permissions in Jira
- Check project permissions
- Verify issue type and workflow configurations

## Contributing

Feel free to submit issues or pull requests to improve these examples!

## License

These examples are part of the jira-cli project and follow the same MIT license.
