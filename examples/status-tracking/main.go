// status-tracking demonstrates how to retrieve and analyze status changes for Jira issues
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/eliziario/jira-lib/lib"
)

func main() {
	// Command line flags
	var (
		server  = flag.String("server", "", "Jira server URL (required)")
		email   = flag.String("email", "", "Email/username for authentication (required)")
		token   = flag.String("token", "", "API token or password (required)")
		issue   = flag.String("issue", "", "Issue key to track (e.g., PROJ-123)")
		project = flag.String("project", "", "Track all issues in project (alternative to -issue)")
		days    = flag.Int("days", 30, "For project tracking, look at issues updated in last N days")
		format  = flag.String("format", "simple", "Output format: simple, detailed, csv, timeline")
		analyze = flag.Bool("analyze", false, "Show analysis of status transitions")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Track status changes for Jira issues.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Track a single issue\n")
		fmt.Fprintf(os.Stderr, "  %s -server https://example.atlassian.net -email user@example.com -token TOKEN -issue PROJ-123\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Track all issues in a project (recent)\n")
		fmt.Fprintf(os.Stderr, "  %s -server ... -project PROJ -days 7\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Generate timeline view\n")
		fmt.Fprintf(os.Stderr, "  %s -server ... -issue PROJ-123 -format timeline\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Analyze transition patterns\n")
		fmt.Fprintf(os.Stderr, "  %s -server ... -project PROJ -analyze\n", os.Args[0])
	}

	flag.Parse()

	// Validate required flags
	if *server == "" || *email == "" || *token == "" {
		fmt.Fprintf(os.Stderr, "Error: server, email, and token are required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if *issue == "" && *project == "" {
		fmt.Fprintf(os.Stderr, "Error: either -issue or -project must be specified\n\n")
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

	if *issue != "" {
		// Track single issue
		trackSingleIssue(client, *issue, *format)
	} else {
		// Track project issues
		trackProjectIssues(client, *project, *days, *format, *analyze)
	}
}

func trackSingleIssue(client *lib.JiraClient, issueKey, format string) {
	fmt.Printf("Fetching status changes for %s...\n\n", issueKey)
	
	changes, err := client.GetIssueStatusChanges(issueKey)
	if err != nil {
		log.Fatalf("Failed to get status changes: %v", err)
	}

	if len(changes) == 0 {
		fmt.Println("No status changes found for this issue.")
		return
	}

	switch format {
	case "timeline":
		printTimeline(issueKey, changes)
	case "detailed":
		printDetailed(issueKey, changes)
	case "csv":
		printCSV([]string{issueKey}, [][]lib.StatusChange{changes})
	default:
		printSimple(issueKey, changes)
	}

	// Show summary statistics
	fmt.Printf("\n--- Summary ---\n")
	fmt.Printf("Total transitions: %d\n", len(changes))
	
	if len(changes) > 0 {
		duration := changes[len(changes)-1].Timestamp.Sub(changes[0].Timestamp)
		fmt.Printf("Time span: %s\n", formatDuration(duration))
		fmt.Printf("Current status: %s\n", changes[len(changes)-1].ToStatus)
		
		// Calculate time in each status
		statusTime := calculateStatusTime(changes)
		fmt.Printf("\nTime in each status:\n")
		for status, duration := range statusTime {
			fmt.Printf("  %-20s %s\n", status+":", formatDuration(duration))
		}
	}
}

func trackProjectIssues(client *lib.JiraClient, project string, days int, format string, analyze bool) {
	fmt.Printf("Fetching recent issues from project %s (last %d days)...\n", project, days)
	
	// Fetch recent issues
	issues, err := client.GetRecentIssues(days, project)
	if err != nil {
		log.Fatalf("Failed to fetch issues: %v", err)
	}

	fmt.Printf("Found %d issues. Fetching status changes...\n", len(issues))
	
	var allIssueKeys []string
	var allChanges [][]lib.StatusChange
	statusCounts := make(map[string]int)
	transitionCounts := make(map[string]int)
	
	for i, issue := range issues {
		if (i+1)%10 == 0 {
			fmt.Printf("  Processing %d/%d...\n", i+1, len(issues))
		}
		
		changes, err := client.GetIssueStatusChanges(issue.Key)
		if err != nil {
			fmt.Printf("  Warning: Failed to get changes for %s: %v\n", issue.Key, err)
			continue
		}
		
		if len(changes) > 0 {
			allIssueKeys = append(allIssueKeys, issue.Key)
			allChanges = append(allChanges, changes)
			
			// Collect statistics
			for _, change := range changes {
				statusCounts[change.ToStatus]++
				if change.FromStatus != "" {
					transition := fmt.Sprintf("%s → %s", change.FromStatus, change.ToStatus)
					transitionCounts[transition]++
				}
			}
		}
	}
	
	fmt.Printf("\nProcessed %d issues with status changes.\n\n", len(allIssueKeys))
	
	if format == "csv" {
		printCSV(allIssueKeys, allChanges)
	} else {
		// Print summary for each issue
		for i, issueKey := range allIssueKeys {
			if format == "detailed" {
				printDetailed(issueKey, allChanges[i])
				fmt.Println(strings.Repeat("-", 80))
			} else {
				printSimple(issueKey, allChanges[i])
			}
		}
	}
	
	if analyze {
		fmt.Printf("\n--- Analysis ---\n")
		fmt.Printf("Total issues with changes: %d\n", len(allIssueKeys))
		
		fmt.Printf("\nMost common statuses:\n")
		for status, count := range statusCounts {
			fmt.Printf("  %-20s %d occurrences\n", status+":", count)
		}
		
		fmt.Printf("\nMost common transitions:\n")
		for transition, count := range transitionCounts {
			if count > 1 {
				fmt.Printf("  %-30s %d times\n", transition+":", count)
			}
		}
		
		// Calculate average time in statuses
		allStatusTimes := make(map[string][]time.Duration)
		for _, changes := range allChanges {
			statusTime := calculateStatusTime(changes)
			for status, duration := range statusTime {
				allStatusTimes[status] = append(allStatusTimes[status], duration)
			}
		}
		
		fmt.Printf("\nAverage time in each status:\n")
		for status, durations := range allStatusTimes {
			var total time.Duration
			for _, d := range durations {
				total += d
			}
			avg := total / time.Duration(len(durations))
			fmt.Printf("  %-20s %s (based on %d issues)\n", status+":", formatDuration(avg), len(durations))
		}
	}
}

func printSimple(issueKey string, changes []lib.StatusChange) {
	fmt.Printf("%s: ", issueKey)
	for i, change := range changes {
		if i > 0 {
			fmt.Print(" → ")
		}
		if change.FromStatus == "" {
			fmt.Printf("%s", change.ToStatus)
		} else {
			fmt.Printf("%s", change.ToStatus)
		}
	}
	fmt.Printf(" (current: %s)\n", changes[len(changes)-1].ToStatus)
}

func printDetailed(issueKey string, changes []lib.StatusChange) {
	fmt.Printf("Issue: %s\n", issueKey)
	fmt.Printf("Status History:\n")
	
	for i, change := range changes {
		author := change.Author
		if author == "" {
			author = "System"
		}
		
		if change.FromStatus == "" {
			fmt.Printf("  %s: Created in '%s' by %s\n",
				change.Timestamp.Format("2006-01-02 15:04"),
				change.ToStatus,
				author)
		} else {
			fmt.Printf("  %s: '%s' → '%s' by %s\n",
				change.Timestamp.Format("2006-01-02 15:04"),
				change.FromStatus,
				change.ToStatus,
				author)
		}
		
		// Show time in status
		if i < len(changes)-1 {
			duration := changes[i+1].Timestamp.Sub(change.Timestamp)
			fmt.Printf("    (Time in '%s': %s)\n", change.ToStatus, formatDuration(duration))
		} else {
			duration := time.Now().Sub(change.Timestamp)
			fmt.Printf("    (Time in '%s': %s - current)\n", change.ToStatus, formatDuration(duration))
		}
	}
}

func printTimeline(issueKey string, changes []lib.StatusChange) {
	fmt.Printf("Timeline for %s:\n\n", issueKey)
	
	maxStatusLen := 0
	for _, change := range changes {
		if len(change.ToStatus) > maxStatusLen {
			maxStatusLen = len(change.ToStatus)
		}
	}
	
	for i, change := range changes {
		// Print date
		fmt.Printf("%s  ", change.Timestamp.Format("2006-01-02"))
		
		// Print status bar
		padding := strings.Repeat(" ", maxStatusLen-len(change.ToStatus))
		fmt.Printf("[%s]%s", change.ToStatus, padding)
		
		// Print transition arrow and duration
		if i < len(changes)-1 {
			duration := changes[i+1].Timestamp.Sub(change.Timestamp)
			fmt.Printf(" ──(%s)──> ", formatDuration(duration))
		} else {
			duration := time.Now().Sub(change.Timestamp)
			fmt.Printf(" ──(%s)──> [Current]", formatDuration(duration))
		}
		
		fmt.Println()
	}
}

func printCSV(issueKeys []string, allChanges [][]lib.StatusChange) {
	fmt.Println("Issue,Timestamp,From Status,To Status,Author,Duration in Status")
	
	for i, issueKey := range issueKeys {
		changes := allChanges[i]
		for j, change := range changes {
			var duration string
			if j < len(changes)-1 {
				d := changes[j+1].Timestamp.Sub(change.Timestamp)
				duration = formatDuration(d)
			} else {
				d := time.Now().Sub(change.Timestamp)
				duration = formatDuration(d) + " (current)"
			}
			
			fmt.Printf("%s,%s,%s,%s,%s,%s\n",
				issueKey,
				change.Timestamp.Format("2006-01-02 15:04:05"),
				escapeCSV(change.FromStatus),
				escapeCSV(change.ToStatus),
				escapeCSV(change.Author),
				duration,
			)
		}
	}
}

func calculateStatusTime(changes []lib.StatusChange) map[string]time.Duration {
	statusTime := make(map[string]time.Duration)
	
	for i, change := range changes {
		var duration time.Duration
		if i < len(changes)-1 {
			duration = changes[i+1].Timestamp.Sub(change.Timestamp)
		} else {
			duration = time.Now().Sub(change.Timestamp)
		}
		statusTime[change.ToStatus] += duration
	}
	
	return statusTime
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	
	if days > 0 {
		if hours > 0 {
			return fmt.Sprintf("%dd %dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	}
	
	if hours > 0 {
		minutes := int(d.Minutes()) % 60
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}
	
	minutes := int(d.Minutes())
	if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	
	return "< 1m"
}

func escapeCSV(s string) string {
	if strings.ContainsAny(s, ",\"\n\r") {
		s = strings.ReplaceAll(s, "\"", "\"\"")
		return "\"" + s + "\""
	}
	return s
}