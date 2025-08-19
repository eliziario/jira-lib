# Status Tracking Example

This example demonstrates how to use the `GetIssueStatusChanges` method to track and analyze status transitions in Jira issues.

## Features

- Track status changes for a single issue
- Track status changes for all issues in a project
- Multiple output formats (simple, detailed, timeline, CSV)
- Status transition analysis
- Time tracking in each status
- Transition pattern analysis

## Building

```bash
go build -o status-tracking main.go
```

## Usage

### Track a Single Issue

```bash
# Simple view
./status-tracking \
  -server https://your-domain.atlassian.net \
  -email your-email@example.com \
  -token YOUR_API_TOKEN \
  -issue PROJ-123

# Timeline view
./status-tracking \
  -server https://your-domain.atlassian.net \
  -email your-email@example.com \
  -token YOUR_API_TOKEN \
  -issue PROJ-123 \
  -format timeline

# Detailed view with timestamps
./status-tracking \
  -server https://your-domain.atlassian.net \
  -email your-email@example.com \
  -token YOUR_API_TOKEN \
  -issue PROJ-123 \
  -format detailed
```

### Track Project Issues

```bash
# Track recent issues in a project
./status-tracking \
  -server https://your-domain.atlassian.net \
  -email your-email@example.com \
  -token YOUR_API_TOKEN \
  -project PROJ \
  -days 7

# Analyze transition patterns
./status-tracking \
  -server https://your-domain.atlassian.net \
  -email your-email@example.com \
  -token YOUR_API_TOKEN \
  -project PROJ \
  -days 30 \
  -analyze

# Export to CSV
./status-tracking \
  -server https://your-domain.atlassian.net \
  -email your-email@example.com \
  -token YOUR_API_TOKEN \
  -project PROJ \
  -format csv > status-changes.csv
```

## Output Formats

### Simple Format
Shows the status progression in a compact format:
```
PROJ-123: To Do → In Progress → Code Review → Testing → Done (current: Done)
```

### Timeline Format
Visual representation of status changes over time:
```
Timeline for PROJ-123:

2024-01-01  [To Do]        ──(2d 3h)──> 
2024-01-03  [In Progress]  ──(1d 5h)──> 
2024-01-04  [Code Review]  ──(4h)──> 
2024-01-05  [Testing]      ──(1d)──> 
2024-01-06  [Done]         ──(10d)──> [Current]
```

### Detailed Format
Complete information about each transition:
```
Issue: PROJ-123
Status History:
  2024-01-01 10:00: Created in 'To Do' by john.doe
    (Time in 'To Do': 2d 3h)
  2024-01-03 13:00: 'To Do' → 'In Progress' by john.doe
    (Time in 'In Progress': 1d 5h)
  2024-01-04 18:00: 'In Progress' → 'Code Review' by jane.smith
    (Time in 'Code Review': 4h)
  ...
```

### CSV Format
Spreadsheet-compatible format for data analysis:
```csv
Issue,Timestamp,From Status,To Status,Author,Duration in Status
PROJ-123,2024-01-01 10:00:00,,To Do,john.doe,2d 3h
PROJ-123,2024-01-03 13:00:00,To Do,In Progress,john.doe,1d 5h
...
```

## Analysis Features

When using the `-analyze` flag with project tracking, the tool provides:

### Transition Statistics
- Most common statuses
- Most common transitions
- Total number of transitions

### Time Analysis
- Average time spent in each status
- Time distribution across statuses
- Identification of bottlenecks

Example output:
```
--- Analysis ---
Total issues with changes: 42

Most common statuses:
  In Progress:         156 occurrences
  Code Review:         98 occurrences
  Done:               85 occurrences

Most common transitions:
  To Do → In Progress:         42 times
  In Progress → Code Review:   38 times
  Code Review → Testing:       35 times

Average time in each status:
  To Do:               3d 5h (based on 42 issues)
  In Progress:         2d 8h (based on 40 issues)
  Code Review:         1d 2h (based on 38 issues)
  Testing:             1d 6h (based on 35 issues)
```

## Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-server` | Jira server URL (required) | - |
| `-email` | Email/username for authentication (required) | - |
| `-token` | API token or password (required) | - |
| `-issue` | Single issue key to track | - |
| `-project` | Project key to track all issues | - |
| `-days` | For project tracking, look at issues updated in last N days | 30 |
| `-format` | Output format: simple, detailed, timeline, csv | simple |
| `-analyze` | Show analysis of status transitions | false |

## Use Cases

### 1. Issue Lifecycle Tracking
Track how long an issue spends in each status to identify bottlenecks:
```bash
./status-tracking -server ... -issue PROJ-123 -format detailed
```

### 2. Team Performance Analysis
Analyze how quickly issues move through your workflow:
```bash
./status-tracking -server ... -project PROJ -days 30 -analyze
```

### 3. SLA Monitoring
Export status changes to CSV for SLA reporting:
```bash
./status-tracking -server ... -project PROJ -format csv > sla-report.csv
```

### 4. Process Improvement
Identify which transitions take the longest:
```bash
./status-tracking -server ... -project PROJ -analyze | grep "Average time"
```

## Notes

- The tool fetches the complete changelog for each issue, including pagination
- Initial status is inferred from the first transition
- Current time is used to calculate duration for issues still in progress
- Large projects with many issues may take some time to process