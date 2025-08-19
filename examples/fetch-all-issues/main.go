// fetch-all-issues demonstrates how to use the GetAllIssues functionality
// to retrieve all issues from a Jira instance with various filtering options.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/eliziario/jira-lib/lib"
	"github.com/eliziario/jira-lib/pkg/jira"
)

func main() {
	// Command line flags
	var (
		server     = flag.String("server", "", "Jira server URL (required)")
		email      = flag.String("email", "", "Email/username for authentication (required)")
		token      = flag.String("token", "", "API token or password (required)")
		project    = flag.String("project", "", "Filter by project key (optional)")
		startDate  = flag.String("start-date", "", "Filter issues created after this date (YYYY-MM-DD)")
		dateField  = flag.String("date-field", "created", "Date field to filter on: created, updated, or resolved")
		maxResults = flag.Int("max", 0, "Maximum number of issues to fetch (0 for unlimited)")
		jql        = flag.String("jql", "", "Additional JQL filter to apply")
		orderBy    = flag.String("order", "", "Order by field (default: created DESC)")
		format     = flag.String("format", "simple", "Output format: simple, detailed, csv")
		days       = flag.Int("days", 0, "Fetch issues from last N days (alternative to start-date)")
		dateRange  = flag.String("date-range", "", "Date range in format START:END (YYYY-MM-DD:YYYY-MM-DD)")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Fetch all issues from Jira with optional filtering.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Fetch all issues from a project\n")
		fmt.Fprintf(os.Stderr, "  %s -server https://example.atlassian.net -email user@example.com -token YOUR_TOKEN -project PROJ\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Fetch issues created in the last 7 days\n")
		fmt.Fprintf(os.Stderr, "  %s -server ... -days 7\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Fetch issues with custom JQL filter\n")
		fmt.Fprintf(os.Stderr, "  %s -server ... -jql \"status = 'In Progress' AND priority = High\"\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Fetch issues in date range\n")
		fmt.Fprintf(os.Stderr, "  %s -server ... -date-range 2024-01-01:2024-01-31\n", os.Args[0])
	}

	flag.Parse()

	// Validate required flags
	if *server == "" || *email == "" || *token == "" {
		fmt.Fprintf(os.Stderr, "Error: server, email, and token are required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Create client
	client, err := lib.NewClient(lib.ClientConfig{
		Server:   *server,
		Login:    *email,
		APIToken: *token,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Determine which fetch method to use
	var issues []*jira.Issue
	startTime := time.Now()

	if *dateRange != "" {
		// Use date range method
		parts := strings.Split(*dateRange, ":")
		if len(parts) != 2 {
			log.Fatalf("Invalid date range format. Use START:END (e.g., 2024-01-01:2024-01-31)")
		}
		fmt.Printf("Fetching issues from %s to %s...\n", parts[0], parts[1])
		issues, err = client.GetIssuesByDateRange(parts[0], parts[1], *dateField)
		if err != nil {
			log.Fatalf("Failed to fetch issues: %v", err)
		}
	} else if *days > 0 {
		// Use recent issues method
		fmt.Printf("Fetching issues from the last %d days...\n", *days)
		issues, err = client.GetRecentIssues(*days, *project)
		if err != nil {
			log.Fatalf("Failed to fetch issues: %v", err)
		}
	} else {
		// Use general GetAllIssues method
		options := lib.GetAllIssuesOptions{
			Project:    *project,
			StartDate:  *startDate,
			DateField:  *dateField,
			MaxResults: *maxResults,
			JQL:        *jql,
			OrderBy:    *orderBy,
		}
		
		fmt.Println("Fetching issues...")
		issues, err = client.GetAllIssues(options)
		if err != nil {
			log.Fatalf("Failed to fetch issues: %v", err)
		}
	}

	elapsed := time.Since(startTime)
	fmt.Printf("Fetched %d issues in %s\n\n", len(issues), elapsed.Round(time.Millisecond))

	// Output results based on format
	switch *format {
	case "csv":
		printCSV(issues)
	case "detailed":
		printDetailed(issues)
	default:
		printSimple(issues)
	}
}

func printSimple(issues []*jira.Issue) {
	if len(issues) == 0 {
		fmt.Println("No issues found.")
		return
	}

	// Print header
	fmt.Printf("%-15s %-12s %-10s %-15s %s\n", "Key", "Type", "Status", "Assignee", "Summary")
	fmt.Println(strings.Repeat("-", 100))

	// Print issues
	for _, issue := range issues {
		assignee := "Unassigned"
		if issue.Fields.Assignee.Name != "" {
			assignee = issue.Fields.Assignee.Name
		}

		issueType := "Unknown"
		if issue.Fields.IssueType.Name != "" {
			issueType = issue.Fields.IssueType.Name
		}

		status := "Unknown"
		if issue.Fields.Status.Name != "" {
			status = issue.Fields.Status.Name
		}

		summary := issue.Fields.Summary
		if len(summary) > 40 {
			summary = summary[:37] + "..."
		}

		fmt.Printf("%-15s %-12s %-10s %-15s %s\n",
			issue.Key,
			truncate(issueType, 12),
			truncate(status, 10),
			truncate(assignee, 15),
			summary,
		)
	}
}

func printDetailed(issues []*jira.Issue) {
	if len(issues) == 0 {
		fmt.Println("No issues found.")
		return
	}

	for i, issue := range issues {
		if i > 0 {
			fmt.Println(strings.Repeat("-", 80))
		}

		fmt.Printf("Issue: %s\n", issue.Key)
		fmt.Printf("Summary: %s\n", issue.Fields.Summary)
		
		if issue.Fields.IssueType.Name != "" {
			fmt.Printf("Type: %s\n", issue.Fields.IssueType.Name)
		}
		
		if issue.Fields.Status.Name != "" {
			fmt.Printf("Status: %s\n", issue.Fields.Status.Name)
		}
		
		if issue.Fields.Priority.Name != "" {
			fmt.Printf("Priority: %s\n", issue.Fields.Priority.Name)
		}
		
		if issue.Fields.Assignee.Name != "" {
			fmt.Printf("Assignee: %s\n", issue.Fields.Assignee.Name)
		} else {
			fmt.Printf("Assignee: Unassigned\n")
		}
		
		if issue.Fields.Reporter.Name != "" {
			fmt.Printf("Reporter: %s\n", issue.Fields.Reporter.Name)
		}
		
		if issue.Fields.Created != "" {
			fmt.Printf("Created: %s\n", formatTime(issue.Fields.Created))
		}
		
		if issue.Fields.Updated != "" {
			fmt.Printf("Updated: %s\n", formatTime(issue.Fields.Updated))
		}
		
		if len(issue.Fields.Labels) > 0 {
			fmt.Printf("Labels: %s\n", strings.Join(issue.Fields.Labels, ", "))
		}
		
		if len(issue.Fields.Components) > 0 {
			var componentNames []string
			for _, c := range issue.Fields.Components {
				componentNames = append(componentNames, c.Name)
			}
			fmt.Printf("Components: %s\n", strings.Join(componentNames, ", "))
		}
		
		// Note: URL would need to be constructed from server config as issue.Self is not available
	}
}

func printCSV(issues []*jira.Issue) {
	if len(issues) == 0 {
		fmt.Println("No issues found.")
		return
	}

	// Print CSV header
	fmt.Println("Key,Type,Status,Priority,Assignee,Reporter,Summary,Created,Updated,Labels")

	// Print issues
	for _, issue := range issues {
		assignee := ""
		if issue.Fields.Assignee.Name != "" {
			assignee = issue.Fields.Assignee.Name
		}

		reporter := ""
		if issue.Fields.Reporter.Name != "" {
			reporter = issue.Fields.Reporter.Name
		}

		issueType := ""
		if issue.Fields.IssueType.Name != "" {
			issueType = issue.Fields.IssueType.Name
		}

		status := ""
		if issue.Fields.Status.Name != "" {
			status = issue.Fields.Status.Name
		}

		priority := ""
		if issue.Fields.Priority.Name != "" {
			priority = issue.Fields.Priority.Name
		}

		labels := strings.Join(issue.Fields.Labels, ";")
		
		// Escape fields that might contain commas
		summary := escapeCSV(issue.Fields.Summary)
		
		fmt.Printf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			issue.Key,
			issueType,
			status,
			priority,
			assignee,
			reporter,
			summary,
			formatTime(issue.Fields.Created),
			formatTime(issue.Fields.Updated),
			labels,
		)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatTime(timeStr string) string {
	if timeStr == "" {
		return ""
	}
	
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		// Try alternative format
		t, err = time.Parse("2006-01-02T15:04:05.000-0700", timeStr)
		if err != nil {
			return timeStr // Return as-is if parsing fails
		}
	}
	
	return t.Format("2006-01-02 15:04")
}

func escapeCSV(s string) string {
	if strings.ContainsAny(s, ",\"\n\r") {
		s = strings.ReplaceAll(s, "\"", "\"\"")
		return "\"" + s + "\""
	}
	return s
}

