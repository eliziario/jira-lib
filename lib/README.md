# JiraCLI Library

This package provides a clean Go library interface for interacting with Jira, based on the jira-cli tool.

## Installation

```bash
go get github.com/eliziario/jira-lib/lib
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/eliziario/jira-lib/lib"
)

func main() {
    // Create a client for Jira Cloud
    client, err := lib.NewClient(lib.ClientConfig{
        Server:   "https://your-domain.atlassian.net",
        Login:    "your-email@example.com",
        APIToken: "your-api-token",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Get an issue
    issue, err := client.GetIssue("PROJ-123")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Issue: %s - %s\n", issue.Key, issue.Fields.Summary)
}
```

## Configuration

### Basic Configuration

```go
config := lib.ClientConfig{
    Server:   "https://your-domain.atlassian.net",
    Login:    "your-email@example.com",
    APIToken: "your-api-token",
}
```

### On-Premise Jira

```go
config := lib.ClientConfig{
    Server:           "https://jira.company.com",
    Login:            "username",
    APIToken:         "password", // Use password for basic auth
    InstallationType: "Local",
}
```

### Using Bearer Token (PAT)

```go
config := lib.ClientConfig{
    Server:   "https://jira.company.com",
    Login:    "username",
    APIToken: "personal-access-token",
    AuthType: "bearer",
}
```

### mTLS Authentication

```go
config := lib.ClientConfig{
    Server:   "https://jira.company.com",
    Login:    "username",
    AuthType: "mtls",
    MTLSConfig: &lib.MTLSConfig{
        CaCert:     "/path/to/ca.crt",
        ClientCert: "/path/to/client.crt",
        ClientKey:  "/path/to/client.key",
    },
}
```

### Advanced Configuration

```go
config := lib.ClientConfig{
    Server:   "https://your-domain.atlassian.net",
    Login:    "your-email@example.com",
    APIToken: "your-api-token",
    
    // Optional settings
    Insecure: true,                     // Allow insecure SSL connections
    Debug:    true,                      // Enable debug logging
    Timeout:  30 * time.Second,          // Custom timeout (default: 15s)
    InstallationType: "Cloud",          // "Cloud" or "Local" (default: "Cloud")
}
```

## Common Operations

### Search Issues

```go
// Search using JQL
results, err := client.SearchIssues("project = PROJ AND status = 'Open'", 0, 50)
if err != nil {
    log.Fatal(err)
}

for _, issue := range results.Issues {
    fmt.Printf("%s: %s\n", issue.Key, issue.Fields.Summary)
}
```

### Create Issue

```go
import "github.com/eliziario/jira-lib/pkg/jira"

createRequest := &jira.CreateRequest{
    Project: "PROJ",
    Name:    "Task",
    Summary: "New task",
    Body: &jira.CreateRequestBody{
        Type:    "doc",
        Version: 1,
        Content: []map[string]interface{}{
            {
                "type": "paragraph",
                "content": []map[string]interface{}{
                    {
                        "type": "text",
                        "text": "Task description",
                    },
                },
            },
        },
    },
    Priority: "High",
    Labels:   []string{"backend", "urgent"},
}

response, err := client.CreateIssue(createRequest)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Created: %s\n", response.Key)
```

### Update Issue

```go
updateRequest := &jira.EditRequest{
    Summary:  "Updated summary",
    Priority: "Low",
    Labels:   []string{"updated"},
}

err := client.UpdateIssue("PROJ-123", updateRequest)
```

### Transition Issue

```go
// Get available transitions
transitions, err := client.GetTransitions("PROJ-123")
if err != nil {
    log.Fatal(err)
}

// Find desired transition
var transitionID string
for _, t := range transitions {
    if t.Name == "Done" {
        transitionID = t.ID
        break
    }
}

// Apply transition
if transitionID != "" {
    err = client.TransitionIssue("PROJ-123", transitionID, &jira.TransitionRequest{
        Transition: &jira.TransitionRequestData{
            ID: transitionID,
        },
    })
}
```

### Add Comment

```go
// Add public comment
err := client.AddComment("PROJ-123", "This is a comment", false)

// Add internal comment
err := client.AddComment("PROJ-123", "Internal note", true)
```

### Work with Projects

```go
// Get all projects
projects, err := client.GetProjects()

// Get specific project
project, err := client.GetProject("PROJ")
```

### Work with Boards and Sprints

```go
// Get boards
boards, err := client.GetBoards("PROJ", "scrum")

// Get sprints
sprints, err := client.GetSprints(boardID, "active", 0, 50)

// Get sprint issues
sprintIssues, err := client.GetSprintIssues(sprintID, "", 0, 50)
```

### Work with Epics

```go
// Get epics
epics, err := client.GetEpics(boardID, 0, 50)

// Get epic issues
issues, total, err := client.GetEpicIssues("EPIC-1", "", 0, 50)
```

## Using the Raw Client

For operations not covered by the wrapper, you can access the underlying client:

```go
rawClient := client.GetRawClient()

// Now you can use any method from pkg/jira
meta, err := rawClient.GetCreateMeta(&jira.CreateMetaRequest{
    Projects: []string{"PROJ"},
    Expand:   "projects.issuetypes.fields",
})
```

## Alternative: Using the API Package

If you prefer more control and don't mind some CLI-specific dependencies, you can use the `api` package directly:

```go
import (
    "github.com/eliziario/jira-lib/api"
    "github.com/eliziario/jira-lib/pkg/jira"
)

// Create client without global state
client := api.NewClient(jira.Config{
    Server:   "https://your-domain.atlassian.net",
    Login:    "your-email@example.com",
    APIToken: "your-api-token",
})

// Use the client
issue, err := client.GetIssue("PROJ-123")
```

## Error Handling

The library uses typed errors from the `pkg/jira` package:

```go
issue, err := client.GetIssue("PROJ-999")
if err != nil {
    switch e := err.(type) {
    case *jira.ErrUnexpectedResponse:
        fmt.Printf("HTTP %d: %s\n", e.StatusCode, e.Status)
        fmt.Printf("Error: %s\n", e.Body)
    default:
        if err == jira.ErrNoResult {
            fmt.Println("Issue not found")
        } else {
            fmt.Printf("Error: %v\n", err)
        }
    }
}
```

## Examples

See the [example_test.go](example_test.go) file for more usage examples.

## License

This library is part of the jira-cli project and follows the same MIT license.