package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/eliziario/jira-lib/lib"
	"github.com/eliziario/jira-lib/pkg/jira"
)

// Config holds the application configuration
type Config struct {
	Server           string `json:"server"`
	Login            string `json:"login"`
	APIToken         string `json:"api_token"`
	Project          string `json:"default_project"`
	InstallationType string `json:"installation_type"`
}

// Application holds the main application state
type Application struct {
	client  *lib.JiraClient
	config  Config
	scanner *bufio.Scanner
}

func main() {
	var (
		configFile = flag.String("config", "", "Path to config file (JSON)")
		debug      = flag.Bool("debug", false, "Enable debug mode")
	)
	flag.Parse()

	app := &Application{
		scanner: bufio.NewScanner(os.Stdin),
	}

	// Load configuration
	if err := app.loadConfig(*configFile); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create Jira client
	clientConfig := lib.ClientConfig{
		Server:           app.config.Server,
		Login:            app.config.Login,
		APIToken:         app.config.APIToken,
		Debug:            *debug,
		InstallationType: app.config.InstallationType,
	}

	client, err := lib.NewClient(clientConfig)
	if err != nil {
		log.Fatalf("Failed to create Jira client: %v", err)
	}
	app.client = client

	// Verify connection
	fmt.Println("Connecting to Jira...")
	if err := app.verifyConnection(); err != nil {
		log.Fatalf("Failed to connect to Jira: %v", err)
	}

	// Run interactive CLI
	app.runInteractiveCLI()
}

func (app *Application) loadConfig(configFile string) error {
	// Try to load from file first
	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		if err := json.Unmarshal(data, &app.config); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
		return nil
	}

	// Try environment variables
	app.config.Server = os.Getenv("JIRA_SERVER")
	app.config.Login = os.Getenv("JIRA_LOGIN")
	app.config.APIToken = os.Getenv("JIRA_API_TOKEN")
	app.config.Project = os.Getenv("JIRA_PROJECT")
	app.config.InstallationType = os.Getenv("JIRA_INSTALLATION_TYPE")

	// Interactive setup if no config found
	if app.config.Server == "" || app.config.Login == "" || app.config.APIToken == "" {
		fmt.Println("No configuration found. Let's set it up!")
		return app.interactiveSetup()
	}

	return nil
}

func (app *Application) interactiveSetup() error {
	fmt.Print("Jira Server URL: ")
	app.scanner.Scan()
	app.config.Server = strings.TrimSpace(app.scanner.Text())

	fmt.Print("Login (email/username): ")
	app.scanner.Scan()
	app.config.Login = strings.TrimSpace(app.scanner.Text())

	fmt.Print("API Token/Password: ")
	app.scanner.Scan()
	app.config.APIToken = strings.TrimSpace(app.scanner.Text())

	fmt.Print("Default Project Key (optional): ")
	app.scanner.Scan()
	app.config.Project = strings.TrimSpace(app.scanner.Text())

	fmt.Print("Installation Type (Cloud/Local) [Cloud]: ")
	app.scanner.Scan()
	app.config.InstallationType = strings.TrimSpace(app.scanner.Text())
	if app.config.InstallationType == "" {
		app.config.InstallationType = "Cloud"
	}

	// Offer to save config
	fmt.Print("\nSave configuration to file? (y/n): ")
	app.scanner.Scan()
	if strings.ToLower(app.scanner.Text()) == "y" {
		return app.saveConfig()
	}

	return nil
}

func (app *Application) saveConfig() error {
	configPath := "jira-config.json"
	data, err := json.MarshalIndent(app.config, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return err
	}

	fmt.Printf("Configuration saved to %s\n", configPath)
	fmt.Println("WARNING: This file contains your API token. Keep it secure!")
	return nil
}

func (app *Application) verifyConnection() error {
	me, err := app.client.GetMyself()
	if err != nil {
		return err
	}

	fmt.Printf("Connected as: %s (%s)\n", me.Name, me.Email)
	fmt.Println()
	return nil
}

func (app *Application) runInteractiveCLI() {
	fmt.Println("Jira CLI - Interactive Mode")
	fmt.Println("Type 'help' for available commands or 'quit' to exit")
	fmt.Println()

	for {
		fmt.Print("> ")
		app.scanner.Scan()
		command := strings.TrimSpace(app.scanner.Text())

		if command == "" {
			continue
		}

		parts := strings.Fields(command)
		cmd := parts[0]
		args := parts[1:]

		switch cmd {
		case "help", "h":
			app.showHelp()
		case "quit", "q", "exit":
			fmt.Println("Goodbye!")
			return
		case "search", "s":
			app.searchIssues(args)
		case "view", "v":
			app.viewIssue(args)
		case "create", "c":
			app.createIssue(args)
		case "update", "u":
			app.updateIssue(args)
		case "comment":
			app.addComment(args)
		case "transition", "t":
			app.transitionIssue(args)
		case "assign", "a":
			app.assignIssue(args)
		case "watch", "w":
			app.watchIssue(args)
		case "projects", "p":
			app.listProjects()
		case "sprint":
			app.sprintOperations(args)
		case "bulk":
			app.bulkOperations(args)
		default:
			fmt.Printf("Unknown command: %s\n", cmd)
			fmt.Println("Type 'help' for available commands")
		}
		fmt.Println()
	}
}

func (app *Application) showHelp() {
	help := `
Available Commands:
  search [query]     - Search issues with JQL (alias: s)
  view <key>         - View issue details (alias: v)
  create             - Create new issue interactively (alias: c)
  update <key>       - Update issue (alias: u)
  comment <key>      - Add comment to issue
  transition <key>   - Change issue status (alias: t)
  assign <key>       - Assign issue to user (alias: a)
  watch <key>        - Watch/unwatch issue (alias: w)
  projects           - List all projects (alias: p)
  sprint             - Sprint operations
  bulk               - Bulk operations
  help               - Show this help (alias: h)
  quit               - Exit the program (alias: q)

Examples:
  search project = PROJ AND status = Open
  view PROJ-123
  assign PROJ-123 john.doe@example.com
  sprint issues 123
`
	fmt.Println(help)
}

func (app *Application) searchIssues(args []string) {
	var jql string
	if len(args) > 0 {
		jql = strings.Join(args, " ")
	} else {
		// Default search for current project
		if app.config.Project != "" {
			jql = fmt.Sprintf("project = %s ORDER BY created DESC", app.config.Project)
		} else {
			fmt.Print("Enter JQL query: ")
			app.scanner.Scan()
			jql = app.scanner.Text()
		}
	}

	fmt.Printf("Searching: %s\n", jql)
	
	results, err := app.client.SearchIssues(jql, 0, 50)
	if err != nil {
		app.handleError("Failed to search issues", err)
		return
	}

	if len(results.Issues) == 0 {
		fmt.Println("No issues found")
		return
	}

	// Display results in table format
	fmt.Printf("\n%-10s %-10s %-15s %-50s\n", "Key", "Type", "Status", "Summary")
	fmt.Println(strings.Repeat("-", 85))

	for _, issue := range results.Issues {
		summary := issue.Fields.Summary
		if len(summary) > 47 {
			summary = summary[:47] + "..."
		}
		fmt.Printf("%-10s %-10s %-15s %-50s\n",
			issue.Key,
			issue.Fields.IssueType.Name,
			issue.Fields.Status.Name,
			summary,
		)
	}
	fmt.Printf("\nTotal: %d issues\n", results.Total)
}

func (app *Application) viewIssue(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: view <issue-key>")
		return
	}

	key := args[0]
	issue, err := app.client.GetIssue(key)
	if err != nil {
		app.handleError("Failed to get issue", err)
		return
	}

	// Display formatted issue
	fmt.Printf("\n%s: %s\n", issue.Key, issue.Fields.Summary)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Type:        %s\n", issue.Fields.IssueType.Name)
	fmt.Printf("Status:      %s\n", issue.Fields.Status.Name)
	fmt.Printf("Priority:    %s\n", issue.Fields.Priority.Name)
	fmt.Printf("Reporter:    %s\n", issue.Fields.Reporter.Name)
	fmt.Printf("Assignee:    %s\n", getAssigneeName(issue.Fields.Assignee))
	fmt.Printf("Created:     %s\n", formatTime(issue.Fields.Created))
	fmt.Printf("Updated:     %s\n", formatTime(issue.Fields.Updated))

	if len(issue.Fields.Labels) > 0 {
		fmt.Printf("Labels:      %s\n", strings.Join(issue.Fields.Labels, ", "))
	}

	if issue.Fields.Description != "" {
		fmt.Printf("\nDescription:\n%s\n", issue.Fields.Description)
	}
}

func (app *Application) createIssue(args []string) {
	project := app.config.Project
	if project == "" {
		fmt.Print("Project key: ")
		app.scanner.Scan()
		project = app.scanner.Text()
	}

	fmt.Print("Issue type (Bug/Task/Story) [Task]: ")
	app.scanner.Scan()
	issueType := app.scanner.Text()
	if issueType == "" {
		issueType = "Task"
	}

	fmt.Print("Summary: ")
	app.scanner.Scan()
	summary := app.scanner.Text()

	fmt.Print("Description: ")
	app.scanner.Scan()
	description := app.scanner.Text()

	fmt.Print("Priority (Highest/High/Medium/Low/Lowest) [Medium]: ")
	app.scanner.Scan()
	priority := app.scanner.Text()
	if priority == "" {
		priority = "Medium"
	}

	request := &jira.CreateRequest{
		Project:  project,
		Name:     issueType,
		Summary:  summary,
		Body:     description,
		Priority: priority,
	}

	response, err := app.client.CreateIssue(request)
	if err != nil {
		app.handleError("Failed to create issue", err)
		return
	}

	fmt.Printf("\nIssue created: %s\n", response.Key)
}

func (app *Application) updateIssue(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: update <issue-key>")
		return
	}

	key := args[0]
	
	// Get current issue
	issue, err := app.client.GetIssue(key)
	if err != nil {
		app.handleError("Failed to get issue", err)
		return
	}

	fmt.Printf("Updating %s: %s\n", issue.Key, issue.Fields.Summary)
	
	request := &jira.EditRequest{}
	
	fmt.Print("New summary (Enter to skip): ")
	app.scanner.Scan()
	if text := app.scanner.Text(); text != "" {
		request.Summary = text
	}

	fmt.Print("New priority (Enter to skip): ")
	app.scanner.Scan()
	if text := app.scanner.Text(); text != "" {
		request.Priority = text
	}

	fmt.Print("Add labels (comma-separated, Enter to skip): ")
	app.scanner.Scan()
	if text := app.scanner.Text(); text != "" {
		request.Labels = strings.Split(text, ",")
		for i := range request.Labels {
			request.Labels[i] = strings.TrimSpace(request.Labels[i])
		}
	}

	err = app.client.UpdateIssue(key, request)
	if err != nil {
		app.handleError("Failed to update issue", err)
		return
	}

	fmt.Println("Issue updated successfully")
}

func (app *Application) addComment(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: comment <issue-key>")
		return
	}

	key := args[0]
	
	fmt.Print("Enter comment: ")
	app.scanner.Scan()
	comment := app.scanner.Text()

	fmt.Print("Internal comment? (y/n) [n]: ")
	app.scanner.Scan()
	internal := strings.ToLower(app.scanner.Text()) == "y"

	err := app.client.AddComment(key, comment, internal)
	if err != nil {
		app.handleError("Failed to add comment", err)
		return
	}

	fmt.Println("Comment added successfully")
}

func (app *Application) transitionIssue(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: transition <issue-key>")
		return
	}

	key := args[0]
	
	transitions, err := app.client.GetTransitions(key)
	if err != nil {
		app.handleError("Failed to get transitions", err)
		return
	}

	if len(transitions) == 0 {
		fmt.Println("No transitions available")
		return
	}

	fmt.Println("Available transitions:")
	for i, t := range transitions {
		fmt.Printf("  %d. %s\n", i+1, t.Name)
	}

	fmt.Print("Select transition: ")
	var choice int
	fmt.Scanln(&choice)

	if choice < 1 || choice > len(transitions) {
		fmt.Println("Invalid choice")
		return
	}

	selected := transitions[choice-1]
	request := &jira.TransitionRequest{
		Transition: &jira.TransitionRequestData{
			ID: string(selected.ID),
		},
	}

	err = app.client.TransitionIssue(key, request)
	if err != nil {
		app.handleError("Failed to transition issue", err)
		return
	}

	fmt.Printf("Issue transitioned to '%s'\n", selected.Name)
}

func (app *Application) assignIssue(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: assign <issue-key> [user]")
		return
	}

	key := args[0]
	var assignee string
	
	if len(args) > 1 {
		assignee = args[1]
	} else {
		fmt.Print("Assignee (email or 'me'): ")
		app.scanner.Scan()
		assignee = app.scanner.Text()
	}

	if assignee == "me" {
		me, err := app.client.GetMyself()
		if err != nil {
			app.handleError("Failed to get current user", err)
			return
		}
		// Use login/email as assignee for 'me'
		assignee = me.Login
	}

	err := app.client.AssignIssue(key, assignee)
	if err != nil {
		app.handleError("Failed to assign issue", err)
		return
	}

	fmt.Println("Issue assigned successfully")
}

func (app *Application) watchIssue(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: watch <issue-key>")
		return
	}

	key := args[0]
	fmt.Printf("Watching issue %s\n", key)
	
	// Note: The watch functionality would need to be added to the lib
	fmt.Println("Watch functionality not yet implemented in library")
}

func (app *Application) listProjects() {
	projects, err := app.client.GetProjects()
	if err != nil {
		app.handleError("Failed to get projects", err)
		return
	}

	fmt.Printf("\n%-10s %-30s %-20s\n", "Key", "Name", "Lead")
	fmt.Println(strings.Repeat("-", 60))

	for _, project := range projects {
		lead := "N/A"
		if project.Lead.Name != "" {
			lead = project.Lead.Name
		}
		fmt.Printf("%-10s %-30s %-20s\n", project.Key, project.Name, lead)
	}
}

func (app *Application) sprintOperations(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: sprint <list|issues> [sprint-id]")
		return
	}

	// Sprint operations would require board ID
	// This is a simplified example
	fmt.Println("Sprint operations require board ID configuration")
}

func (app *Application) bulkOperations(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: bulk <assign|transition|update> <issue-keys...>")
		return
	}

	operation := args[0]
	if len(args) < 2 {
		fmt.Println("Please provide issue keys")
		return
	}

	keys := args[1:]
	
	switch operation {
	case "assign":
		fmt.Print("Assignee for all issues: ")
		app.scanner.Scan()
		assignee := app.scanner.Text()
		
		for _, key := range keys {
			fmt.Printf("Assigning %s... ", key)
			if err := app.client.AssignIssue(key, assignee); err != nil {
				fmt.Printf("failed: %v\n", err)
			} else {
				fmt.Println("done")
			}
		}
		
	case "transition":
		fmt.Print("Target status: ")
		app.scanner.Scan()
		status := app.scanner.Text()
		
		for _, key := range keys {
			fmt.Printf("Transitioning %s... ", key)
			// Get transitions and find matching one
			transitions, err := app.client.GetTransitions(key)
			if err != nil {
				fmt.Printf("failed: %v\n", err)
				continue
			}
			
			var found bool
			for _, t := range transitions {
				if strings.EqualFold(t.Name, status) {
					req := &jira.TransitionRequest{
						Transition: &jira.TransitionRequestData{
							ID: string(t.ID),
						},
					}
					if err := app.client.TransitionIssue(key, req); err != nil {
						fmt.Printf("failed: %v\n", err)
					} else {
						fmt.Println("done")
						found = true
					}
					break
				}
			}
			if !found {
				fmt.Printf("transition '%s' not found\n", status)
			}
		}
		
	default:
		fmt.Printf("Unknown bulk operation: %s\n", operation)
	}
}

func (app *Application) handleError(context string, err error) {
	// Type assert to get more detailed error info
	if jiraErr, ok := err.(*jira.ErrUnexpectedResponse); ok {
		fmt.Printf("Error: %s\n", context)
		fmt.Printf("Status: %s\n", jiraErr.Status)
		fmt.Printf("Details: %s\n", jiraErr.Body.String())
	} else {
		fmt.Printf("Error: %s: %v\n", context, err)
	}
}

// Helper functions
func getPriority(p struct{ Name string `json:"name"` }) string {
	if p.Name == "" {
		return "None"
	}
	return p.Name
}

func getAssigneeName(assignee struct{ Name string `json:"displayName"` }) string {
	if assignee.Name == "" {
		return "Unassigned"
	}
	return assignee.Name
}

func formatTime(t string) string {
	parsed, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return t
	}
	return parsed.Format("2006-01-02 15:04")
}