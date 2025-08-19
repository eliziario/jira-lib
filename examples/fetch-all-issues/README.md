# Fetch All Issues Example

This example demonstrates how to use the `GetAllIssues`, `GetIssuesByDateRange`, and `GetRecentIssues` methods to fetch issues from Jira with automatic pagination.

## Features

- Fetch all issues from a project or entire instance
- Filter by date (created, updated, or resolved)
- Apply custom JQL filters
- Multiple output formats (simple table, detailed, CSV)
- Automatic pagination handling
- Date range filtering
- Recent issues (last N days)

## Building

```bash
go build -o fetch-all-issues main.go
```

## Usage

### Basic Usage

```bash
# Fetch all issues from a project
./fetch-all-issues \
  -server https://your-domain.atlassian.net \
  -email your-email@example.com \
  -token YOUR_API_TOKEN \
  -project PROJ

# Fetch issues created in the last 7 days
./fetch-all-issues \
  -server https://your-domain.atlassian.net \
  -email your-email@example.com \
  -token YOUR_API_TOKEN \
  -days 7

# Fetch issues with a specific status
./fetch-all-issues \
  -server https://your-domain.atlassian.net \
  -email your-email@example.com \
  -token YOUR_API_TOKEN \
  -jql "status = 'In Progress'"
```

### Advanced Filtering

```bash
# Fetch issues created after a specific date
./fetch-all-issues \
  -server https://your-domain.atlassian.net \
  -email your-email@example.com \
  -token YOUR_API_TOKEN \
  -start-date 2024-01-01

# Fetch issues updated in a date range
./fetch-all-issues \
  -server https://your-domain.atlassian.net \
  -email your-email@example.com \
  -token YOUR_API_TOKEN \
  -date-range 2024-01-01:2024-01-31 \
  -date-field updated

# Combine multiple filters
./fetch-all-issues \
  -server https://your-domain.atlassian.net \
  -email your-email@example.com \
  -token YOUR_API_TOKEN \
  -project PROJ \
  -start-date 2024-01-01 \
  -jql "priority = High AND assignee = currentUser()" \
  -max 100
```

### Output Formats

```bash
# Simple table format (default)
./fetch-all-issues ... -format simple

# Detailed format with all fields
./fetch-all-issues ... -format detailed

# CSV format for spreadsheet import
./fetch-all-issues ... -format csv > issues.csv
```

## Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-server` | Jira server URL (required) | - |
| `-email` | Email/username for authentication (required) | - |
| `-token` | API token or password (required) | - |
| `-project` | Filter by project key | - |
| `-start-date` | Filter issues created after this date (YYYY-MM-DD) | - |
| `-date-field` | Date field to filter on: created, updated, or resolved | created |
| `-max` | Maximum number of issues to fetch (0 for unlimited) | 0 |
| `-jql` | Additional JQL filter to apply | - |
| `-order` | Order by field | created DESC |
| `-format` | Output format: simple, detailed, csv | simple |
| `-days` | Fetch issues from last N days | - |
| `-date-range` | Date range in format START:END (YYYY-MM-DD:YYYY-MM-DD) | - |

## Examples of JQL Filters

The `-jql` flag accepts any valid JQL query. Here are some useful examples:

```bash
# High priority bugs
-jql "priority = High AND issuetype = Bug"

# Issues assigned to current user
-jql "assignee = currentUser()"

# Unresolved issues with specific label
-jql "resolution = Unresolved AND labels = 'backend'"

# Issues in specific status
-jql "status IN ('In Progress', 'In Review')"

# Issues with attachments
-jql "attachments IS NOT EMPTY"

# Issues updated by specific user
-jql "updatedBy = 'john.doe'"
```

## Performance Notes

- The library automatically handles pagination, fetching issues in batches of 100
- For large result sets, consider using the `-max` flag to limit results
- The `-order` flag can significantly impact performance for large datasets
- Date filtering at the JQL level is more efficient than post-processing

## Authentication

This example supports the same authentication methods as the main library:

- **Jira Cloud**: Use email and API token
- **Jira Server**: Use username and password
- **Bearer Token**: Set `-auth-type bearer` (not shown in this simplified example)

## Error Handling

The example includes basic error handling:
- Validates required command-line arguments
- Reports connection errors
- Handles pagination errors gracefully
- Provides clear error messages for invalid date formats