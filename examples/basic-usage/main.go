package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/eliziario/jira-lib/lib"
	"github.com/eliziario/jira-lib/pkg/jira"
)

func main() {
	// Command line flags
	var (
		server   = flag.String("server", "", "Jira server URL (e.g., https://your-domain.atlassian.net)")
		login    = flag.String("login", "", "Your Jira login email/username")
		token    = flag.String("token", "", "Your Jira API token or password")
		project  = flag.String("project", "", "Jira project key (e.g., PROJ)")
		action   = flag.String("action", "list", "Action to perform: list, create, view, update, comment, transition")
		issueKey = flag.String("issue", "", "Issue key for view/update/comment/transition actions")
	)
	flag.Parse()

	// Check for environment variables if flags are not provided
	if *server == "" {
		*server = os.Getenv("JIRA_SERVER")
	}
	if *login == "" {
		*login = os.Getenv("JIRA_LOGIN")
	}
	if *token == "" {
		*token = os.Getenv("JIRA_API_TOKEN")
	}
	if *project == "" {
		*project = os.Getenv("JIRA_PROJECT")
	}

	// Validate required parameters
	if *server == "" || *login == "" || *token == "" {
		fmt.Println("Usage: go run main.go -server=<url> -login=<email> -token=<token> -project=<key> [-action=<action>] [-issue=<key>]")
		fmt.Println("\nYou can also set environment variables:")
		fmt.Println("  JIRA_SERVER, JIRA_LOGIN, JIRA_API_TOKEN, JIRA_PROJECT")
		fmt.Println("\nAvailable actions:")
		fmt.Println("  list       - List recent issues (default)")
		fmt.Println("  create     - Create a new issue")
		fmt.Println("  view       - View an issue (requires -issue)")
		fmt.Println("  update     - Update an issue (requires -issue)")
		fmt.Println("  comment    - Add comment to an issue (requires -issue)")
		fmt.Println("  transition - Transition an issue (requires -issue)")
		os.Exit(1)
	}

	// Create Jira client
	client, err := lib.NewClient(lib.ClientConfig{
		Server:   *server,
		Login:    *login,
		APIToken: *token,
	})
	if err != nil {
		log.Fatalf("Failed to create Jira client: %v", err)
	}

	// Perform the requested action
	switch *action {
	case "list":
		listIssues(client, *project)
	case "create":
		createIssue(client, *project)
	case "view":
		if *issueKey == "" {
			log.Fatal("Issue key is required for view action")
		}
		viewIssue(client, *issueKey)
	case "update":
		if *issueKey == "" {
			log.Fatal("Issue key is required for update action")
		}
		updateIssue(client, *issueKey)
	case "comment":
		if *issueKey == "" {
			log.Fatal("Issue key is required for comment action")
		}
		addComment(client, *issueKey)
	case "transition":
		if *issueKey == "" {
			log.Fatal("Issue key is required for transition action")
		}
		transitionIssue(client, *issueKey)
	default:
		log.Fatalf("Unknown action: %s", *action)
	}
}

func listIssues(client *lib.JiraClient, project string) {
	fmt.Printf("Listing issues for project %s...\n\n", project)

	// Build JQL query
	jql := fmt.Sprintf("project = %s ORDER BY created DESC", project)
	
	// Search for issues
	results, err := client.SearchIssues(jql, 0, 20)
	if err != nil {
		log.Fatalf("Failed to search issues: %v", err)
	}

	if len(results.Issues) == 0 {
		fmt.Println("No issues found")
		return
	}

	// Display issues in a table format
	fmt.Printf("%-10s %-10s %-15s %-50s %-20s\n", "Key", "Type", "Status", "Summary", "Assignee")
	fmt.Println(strings.Repeat("-", 115))

	for _, issue := range results.Issues {
		assignee := "Unassigned"
		if issue.Fields.Assignee.Name != "" {
			assignee = issue.Fields.Assignee.Name
		}
		
		summary := issue.Fields.Summary
		if len(summary) > 47 {
			summary = summary[:47] + "..."
		}

		fmt.Printf("%-10s %-10s %-15s %-50s %-20s\n",
			issue.Key,
			issue.Fields.IssueType.Name,
			issue.Fields.Status.Name,
			summary,
			assignee,
		)
	}

	fmt.Printf("\nTotal: %d issues\n", results.Total)
}

func createIssue(client *lib.JiraClient, project string) {
	fmt.Println("Creating a new issue...")

	// Get input from user
	var summary, description, issueType string
	
	fmt.Print("Issue type (Bug/Task/Story) [Task]: ")
	fmt.Scanln(&issueType)
	if issueType == "" {
		issueType = "Task"
	}

	fmt.Print("Summary: ")
	fmt.Scanln(&summary)
	if summary == "" {
		log.Fatal("Summary is required")
	}

	fmt.Print("Description (optional): ")
	fmt.Scanln(&description)

	// Create the issue
	request := &jira.CreateRequest{
		Project: project,
		Name:    issueType,
		Summary: summary,
		Body:    description,
	}

	response, err := client.CreateIssue(request)
	if err != nil {
		log.Fatalf("Failed to create issue: %v", err)
	}

	fmt.Printf("\nIssue created successfully!\n")
	fmt.Printf("Key: %s\n", response.Key)
	// Note: Server URL would need to be passed in or stored separately
	// fmt.Printf("URL: <server>/browse/%s\n", response.Key)
}

func viewIssue(client *lib.JiraClient, key string) {
	fmt.Printf("Fetching issue %s...\n\n", key)

	issue, err := client.GetIssue(key)
	if err != nil {
		log.Fatalf("Failed to get issue: %v", err)
	}

	// Display issue details
	fmt.Printf("Key:         %s\n", issue.Key)
	fmt.Printf("Summary:     %s\n", issue.Fields.Summary)
	fmt.Printf("Type:        %s\n", issue.Fields.IssueType.Name)
	fmt.Printf("Status:      %s\n", issue.Fields.Status.Name)
	fmt.Printf("Priority:    %s\n", issue.Fields.Priority.Name)
	fmt.Printf("Reporter:    %s\n", issue.Fields.Reporter.Name)
	fmt.Printf("Assignee:    %s\n", getAssigneeName(issue.Fields.Assignee))
	fmt.Printf("Created:     %s\n", issue.Fields.Created)
	fmt.Printf("Updated:     %s\n", issue.Fields.Updated)
	
	if len(issue.Fields.Labels) > 0 {
		fmt.Printf("Labels:      %s\n", strings.Join(issue.Fields.Labels, ", "))
	}
	
	if issue.Fields.Description != "" {
		fmt.Printf("\nDescription:\n%s\n", issue.Fields.Description)
	}

	// Show comments if any
	if issue.Fields.Comment.Total > 0 {
		fmt.Printf("\nComments (%d):\n", issue.Fields.Comment.Total)
		fmt.Println(strings.Repeat("-", 50))
		for _, comment := range issue.Fields.Comment.Comments {
			author := "Unknown"
			if comment.Author.DisplayName != "" {
				author = comment.Author.DisplayName
			} else if comment.Author.Name != "" {
				author = comment.Author.Name
			}
			fmt.Printf("[%s] %s:\n%s\n\n", 
				comment.Created, 
				author,
				comment.Body,
			)
		}
	}
}

func updateIssue(client *lib.JiraClient, key string) {
	fmt.Printf("Updating issue %s...\n", key)

	// Get current issue first
	issue, err := client.GetIssue(key)
	if err != nil {
		log.Fatalf("Failed to get issue: %v", err)
	}

	fmt.Printf("Current summary: %s\n", issue.Fields.Summary)
	
	var newSummary, newPriority string
	
	fmt.Print("New summary (press Enter to keep current): ")
	fmt.Scanln(&newSummary)
	
	fmt.Print("New priority (High/Medium/Low, press Enter to keep current): ")
	fmt.Scanln(&newPriority)

	// Build update request
	request := &jira.EditRequest{}
	
	if newSummary != "" {
		request.Summary = newSummary
	}
	
	if newPriority != "" {
		request.Priority = newPriority
	}

	// Only update if there are changes
	if newSummary == "" && newPriority == "" {
		fmt.Println("No changes to make")
		return
	}

	err = client.UpdateIssue(key, request)
	if err != nil {
		log.Fatalf("Failed to update issue: %v", err)
	}

	fmt.Println("Issue updated successfully!")
}

func addComment(client *lib.JiraClient, key string) {
	fmt.Printf("Adding comment to issue %s...\n", key)

	var comment string
	fmt.Print("Enter comment: ")
	fmt.Scanln(&comment)
	
	if comment == "" {
		log.Fatal("Comment cannot be empty")
	}

	err := client.AddComment(key, comment, false)
	if err != nil {
		log.Fatalf("Failed to add comment: %v", err)
	}

	fmt.Println("Comment added successfully!")
}

func transitionIssue(client *lib.JiraClient, key string) {
	fmt.Printf("Transitioning issue %s...\n\n", key)

	// Get available transitions
	transitions, err := client.GetTransitions(key)
	if err != nil {
		log.Fatalf("Failed to get transitions: %v", err)
	}

	if len(transitions) == 0 {
		fmt.Println("No transitions available for this issue")
		return
	}

	// Display available transitions
	fmt.Println("Available transitions:")
	for i, t := range transitions {
		fmt.Printf("%d. %s\n", i+1, t.Name)
	}

	// Get user choice
	var choice int
	fmt.Print("\nSelect transition (enter number): ")
	fmt.Scanln(&choice)

	if choice < 1 || choice > len(transitions) {
		log.Fatal("Invalid choice")
	}

	selected := transitions[choice-1]

	// Perform transition
	request := &jira.TransitionRequest{
		Transition: &jira.TransitionRequestData{
			ID: string(selected.ID),
		},
	}

	err = client.TransitionIssue(key, request)
	if err != nil {
		log.Fatalf("Failed to transition issue: %v", err)
	}

	fmt.Printf("Issue transitioned to '%s' successfully!\n", selected.Name)
}

// Helper function
func getAssigneeName(assignee struct{ Name string `json:"displayName"` }) string {
	if assignee.Name == "" {
		return "Unassigned"
	}
	return assignee.Name
}