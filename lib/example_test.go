package lib_test

import (
	"fmt"
	"log"

	"github.com/eliziario/jira-lib/lib"
	"github.com/eliziario/jira-lib/pkg/jira"
)

func ExampleNewClient() {
	// Create a new client
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

func ExampleJiraClient_SearchIssues() {
	client, err := lib.NewClient(lib.ClientConfig{
		Server:   "https://your-domain.atlassian.net",
		Login:    "your-email@example.com",
		APIToken: "your-api-token",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Search for issues using JQL
	results, err := client.SearchIssues("project = PROJ AND status = 'In Progress'", 0, 50)
	if err != nil {
		log.Fatal(err)
	}

	for _, issue := range results.Issues {
		fmt.Printf("%s: %s\n", issue.Key, issue.Fields.Summary)
	}
}

func ExampleJiraClient_CreateIssue() {
	client, err := lib.NewClient(lib.ClientConfig{
		Server:   "https://your-domain.atlassian.net",
		Login:    "your-email@example.com",
		APIToken: "your-api-token",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create a new issue
	createRequest := &jira.CreateRequest{
		Project: "PROJ",
		Name:    "Task",
		Summary: "New task from library",
		Body:    "This is the issue description", // Can be string or ADF format
	}

	response, err := client.CreateIssue(createRequest)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created issue: %s\n", response.Key)
}

func ExampleJiraClient_TransitionIssue() {
	client, err := lib.NewClient(lib.ClientConfig{
		Server:   "https://your-domain.atlassian.net",
		Login:    "your-email@example.com",
		APIToken: "your-api-token",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Get available transitions for an issue
	transitions, err := client.GetTransitions("PROJ-123")
	if err != nil {
		log.Fatal(err)
	}

	// Find the "In Progress" transition
	var inProgressID string
	for _, t := range transitions {
		if t.Name == "In Progress" {
			inProgressID = string(t.ID)
			break
		}
	}

	// Transition the issue
	if inProgressID != "" {
		err = client.TransitionIssue("PROJ-123", &jira.TransitionRequest{
			Transition: &jira.TransitionRequestData{
				ID: inProgressID,
			},
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Issue transitioned to In Progress")
	}
}

func ExampleJiraClient_AddComment() {
	client, err := lib.NewClient(lib.ClientConfig{
		Server:   "https://your-domain.atlassian.net",
		Login:    "your-email@example.com",
		APIToken: "your-api-token",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Add a public comment
	err = client.AddComment("PROJ-123", "This is a comment from the library", false)
	if err != nil {
		log.Fatal(err)
	}

	// Add an internal comment (visible only to certain groups)
	err = client.AddComment("PROJ-123", "Internal note", true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Comments added successfully")
}

func ExampleJiraClient_GetRawClient() {
	client, err := lib.NewClient(lib.ClientConfig{
		Server:           "https://your-domain.atlassian.net",
		Login:            "your-email@example.com",
		APIToken:         "your-api-token",
		InstallationType: "Local", // For on-premise Jira
	})
	if err != nil {
		log.Fatal(err)
	}

	// Access the raw client for advanced operations
	rawClient := client.GetRawClient()

	// Use methods not exposed by the wrapper
	fields, err := rawClient.GetCreateMeta(&jira.CreateMetaRequest{
		Projects: "PROJ",
		Expand:   "projects.issuetypes.fields",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d fields\n", len(fields.Projects))
}