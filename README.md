# Jira Go Library

A powerful Go library for interacting with Atlassian Jira Cloud and Server/Data Center APIs.

> **This library builds upon the excellent [jira-cli](https://github.com/ankitpokhrel/jira-cli) project by [Ankit Pokhrel](https://github.com/ankitpokhrel), transforming it from a CLI tool into a reusable Go library.**

## Project Background

This library started with the [jira-cli](https://github.com/ankitpokhrel/jira-cli) project - a feature-rich, interactive command-line tool for Jira that has helped thousands of developers. We recognized that the core Jira API implementation within jira-cli was excellent and could benefit the broader Go community if made available as a standalone library.

### What We've Done

While the core Jira API logic comes from jira-cli, we've made significant enhancements to transform it into a proper library:

#### 🔧 **New Library Interface** (`lib/` package)
- Created a completely new, simplified API wrapper
- Designed clean `ClientConfig` structure for easy initialization  
- Added high-level methods that hide complexity
- Implemented proper error propagation and handling
- Removed all CLI-specific logic and dependencies

#### 🏗️ **Refactored API Package**
- Added `NewClient()` function for stateless client creation
- Removed global state and singleton patterns
- Made the package suitable for concurrent use
- Maintained backward compatibility while adding new features

#### 📚 **Comprehensive Documentation**
- Created extensive library-focused documentation
- Added complete working examples (`examples/` directory)
- Wrote inline code documentation for all public APIs
- Included authentication guides for all supported methods

#### ✨ **Library-Specific Enhancements**
- Unified API for both Cloud and Server installations
- Simplified method signatures for common operations
- Added convenient helper methods
- Proper Go module structure without CLI dependencies
- Clean separation of concerns

#### 🧹 **Major Cleanup**
- Removed ~15,000 lines of CLI/TUI code
- Eliminated 20+ CLI-specific dependencies
- Restructured packages for library use
- Removed vendor lock-in

## Why This Library?

There are scenarios where you need programmatic access to Jira without CLI overhead:

- **Building web services** that integrate with Jira
- **Creating automation scripts** in Go
- **Developing custom tools** with specific workflows
- **Embedding Jira functionality** in existing applications
- **Building company-specific CLIs** with custom business logic

This library provides a clean, focused API for these use cases while preserving the battle-tested Jira integration code from the original project.

## Credits

- **Original jira-cli project**: [Ankit Pokhrel](https://github.com/ankitpokhrel) and [contributors](https://github.com/ankitpokhrel/jira-cli/graphs/contributors) for the core Jira API implementation
- **Library transformation**: Significant refactoring and new development to create this library interface

If you need a full-featured CLI tool for Jira, the original [jira-cli](https://github.com/ankitpokhrel/jira-cli) is excellent!

## Features

- 🚀 **Full API Coverage** - Support for Issues, Projects, Boards, Sprints, Epics, and more
- 🔐 **Multiple Authentication** - API tokens, Basic auth, Bearer tokens (PAT), and mTLS
- ☁️ **Cloud & Server Support** - Works with both Jira Cloud and Server/Data Center
- 🎯 **Type Safety** - Strongly typed Go structs for all Jira entities
- 🛡️ **Error Handling** - Comprehensive error types with detailed information
- 📦 **Clean Library Design** - No CLI dependencies, pure library functionality
- 🔄 **Stateless Clients** - Safe for concurrent use in services and applications

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
    // Create a client
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

## Authentication

### Jira Cloud

```go
client, err := lib.NewClient(lib.ClientConfig{
    Server:   "https://your-domain.atlassian.net",
    Login:    "your-email@example.com",
    APIToken: "your-api-token", // Get from: https://id.atlassian.com/manage-profile/security/api-tokens
})
```

### Jira Server/Data Center

```go
client, err := lib.NewClient(lib.ClientConfig{
    Server:           "https://jira.company.com",
    Login:            "username",
    APIToken:         "password", // Use password for basic auth
    InstallationType: "Local",
})
```

### Personal Access Token (PAT)

```go
client, err := lib.NewClient(lib.ClientConfig{
    Server:   "https://jira.company.com",
    Login:    "username",
    APIToken: "personal-access-token",
    AuthType: "bearer",
})
```

### mTLS Authentication

```go
client, err := lib.NewClient(lib.ClientConfig{
    Server:   "https://jira.company.com",
    Login:    "username",
    AuthType: "mtls",
    MTLSConfig: &lib.MTLSConfig{
        CaCert:     "/path/to/ca.crt",
        ClientCert: "/path/to/client.crt",
        ClientKey:  "/path/to/client.key",
    },
})
```

## Core Operations

### Search Issues

```go
// Search with JQL
results, err := client.SearchIssues("project = PROJ AND status = 'In Progress'", 0, 50)
for _, issue := range results.Issues {
    fmt.Printf("%s: %s\n", issue.Key, issue.Fields.Summary)
}
```

### Create Issue

```go
request := &jira.CreateRequest{
    Project:  "PROJ",
    Name:     "Task",
    Summary:  "New task",
    Body:     "Task description",
    Priority: "High",
    Labels:   []string{"backend", "urgent"},
}

response, err := client.CreateIssue(request)
fmt.Printf("Created: %s\n", response.Key)
```

### Update Issue

```go
request := &jira.EditRequest{
    Summary:  "Updated summary",
    Priority: "Low",
    Labels:   []string{"updated"},
}

err := client.UpdateIssue("PROJ-123", request)
```

### Transition Issue

```go
// Get available transitions
transitions, err := client.GetTransitions("PROJ-123")

// Apply transition
request := &jira.TransitionRequest{
    Transition: &jira.TransitionRequestData{
        ID: transitionID,
    },
}
err = client.TransitionIssue("PROJ-123", request)
```

### Add Comment

```go
// Add public comment
err := client.AddComment("PROJ-123", "This is a comment", false)

// Add internal comment
err := client.AddComment("PROJ-123", "Internal note", true)
```

## Advanced Usage

### Using the Raw Client

For operations not covered by the high-level API, you can access the underlying client:

```go
rawClient := client.GetRawClient()

// Use any method from pkg/jira
meta, err := rawClient.GetCreateMeta(&jira.CreateMetaRequest{
    Projects: "PROJ",
    Expand:   "projects.issuetypes.fields",
})
```

### Using the API Package

For more control over client configuration:

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
```

### Error Handling

The library provides typed errors for better error handling:

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

## Package Structure

- `lib/` - **New** high-level library interface with simplified API
- `api/` - **Enhanced** client initialization and proxy functions
- `pkg/jira/` - Core Jira API client implementation (from jira-cli)
- `pkg/adf/` - Atlassian Document Format utilities
- `pkg/md/` - Markdown conversion utilities
- `pkg/jql/` - JQL query builder
- `pkg/netrc/` - .netrc file support for authentication
- `examples/` - **New** complete working examples demonstrating library usage

## Examples

Check the [examples directory](examples/) for complete working examples:

- **basic-usage** - Simple command-line tool demonstrating core operations
- **advanced-cli** - Full-featured interactive CLI with configuration management

Both examples were created specifically to demonstrate library usage patterns.

## Supported Jira Features

- ✅ Issues (Create, Read, Update, Delete, Transition)
- ✅ Comments (Add, Update, Delete)
- ✅ Projects (List, Get details)
- ✅ Boards (List, Get board configuration)
- ✅ Sprints (List, Get issues in sprint)
- ✅ Epics (List, Get epic issues)
- ✅ Users (Search, Get user details)
- ✅ Worklogs (Add, Update)
- ✅ Issue Links (Create, Delete)
- ✅ Attachments (via raw client)
- ✅ Custom Fields (via raw client)
- ✅ JQL Search

## Requirements

- Go 1.21 or higher
- Jira Cloud or Server/Data Center instance
- API token or appropriate credentials

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

For bugs or issues related to the core Jira API functionality, you may also want to check the original [jira-cli](https://github.com/ankitpokhrel/jira-cli) project.

## License

MIT License - see LICENSE file for details.

This library inherits its MIT license from the original [jira-cli](https://github.com/ankitpokhrel/jira-cli) project.

## Related Projects

- **[jira-cli](https://github.com/ankitpokhrel/jira-cli)** - The original CLI tool this library was derived from
- **[go-jira](https://github.com/andygrunwald/go-jira)** - Another popular Go client library for Atlassian Jira
- **[jira-terminal](https://github.com/mk-5/jira-terminal)** - Terminal UI for Jira

## Support

If you find this library useful:
- ⭐ Star this repository
- 🐛 Report bugs or request features via issues
- 💻 Contribute improvements via pull requests
- ⭐ Also consider starring the original [jira-cli](https://github.com/ankitpokhrel/jira-cli) project