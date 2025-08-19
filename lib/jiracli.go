// Package lib provides a clean library interface for using jira-cli functionality
// in other Go applications.
package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/eliziario/jira-lib/pkg/jira"
	"github.com/eliziario/jira-lib/pkg/jira/filter"
)

// ClientConfig holds the configuration for creating a Jira client.
type ClientConfig struct {
	// Server is the base URL of your Jira instance (required)
	Server string
	
	// Login is the username or email for authentication (required)
	Login string
	
	// APIToken is the API token or password for authentication (required)
	APIToken string
	
	// AuthType specifies the authentication type (optional, defaults to "basic")
	// Possible values: "basic", "bearer", "mtls"
	AuthType string
	
	// Insecure allows connections to servers with invalid certificates (optional)
	Insecure bool
	
	// Debug enables debug logging (optional)
	Debug bool
	
	// Timeout specifies the HTTP client timeout (optional, defaults to 15s)
	Timeout time.Duration
	
	// InstallationType specifies if it's "Cloud" or "Local" (optional, defaults to "Cloud")
	InstallationType string
	
	// MTLSConfig holds mTLS configuration if AuthType is "mtls"
	MTLSConfig *MTLSConfig
}

// MTLSConfig holds mTLS authentication configuration.
type MTLSConfig struct {
	CaCert     string
	ClientCert string
	ClientKey  string
}

// JiraClient wraps the underlying jira.Client with convenience methods.
type JiraClient struct {
	client           *jira.Client
	installationType string
}

// NewClient creates a new Jira client for library usage.
func NewClient(config ClientConfig) (*JiraClient, error) {
	if config.Server == "" {
		return nil, fmt.Errorf("server URL is required")
	}
	if config.Login == "" {
		return nil, fmt.Errorf("login is required")
	}
	if config.APIToken == "" {
		return nil, fmt.Errorf("API token is required")
	}
	
	// Set defaults
	if config.AuthType == "" {
		config.AuthType = "basic"
	}
	if config.Timeout == 0 {
		config.Timeout = 15 * time.Second
	}
	if config.InstallationType == "" {
		config.InstallationType = jira.InstallationTypeCloud
	}
	
	authType := jira.AuthType(config.AuthType)
	jiraConfig := jira.Config{
		Server:   config.Server,
		Login:    config.Login,
		APIToken: config.APIToken,
		AuthType: &authType,
		Insecure: &config.Insecure,
		Debug:    config.Debug,
	}
	
	// Add mTLS config if provided
	if config.MTLSConfig != nil {
		jiraConfig.MTLSConfig = jira.MTLSConfig{
			CaCert:     config.MTLSConfig.CaCert,
			ClientCert: config.MTLSConfig.ClientCert,
			ClientKey:  config.MTLSConfig.ClientKey,
		}
	}
	
	client := jira.NewClient(
		jiraConfig,
		jira.WithTimeout(config.Timeout),
		jira.WithInsecureTLS(config.Insecure),
	)
	
	return &JiraClient{
		client:           client,
		installationType: config.InstallationType,
	}, nil
}

// GetIssue retrieves a single issue by key.
func (c *JiraClient) GetIssue(key string, opts ...filter.Filter) (*jira.Issue, error) {
	if c.installationType == jira.InstallationTypeLocal {
		return c.client.GetIssueV2(key, opts...)
	}
	return c.client.GetIssue(key, opts...)
}

// SearchIssues searches for issues using JQL.
func (c *JiraClient) SearchIssues(jql string, from, limit uint) (*jira.SearchResult, error) {
	if c.installationType == jira.InstallationTypeLocal {
		return c.client.SearchV2(jql, from, limit)
	}
	return c.client.Search(jql, from, limit)
}

// CreateIssue creates a new issue.
func (c *JiraClient) CreateIssue(request *jira.CreateRequest) (*jira.CreateResponse, error) {
	if c.installationType == jira.InstallationTypeLocal {
		return c.client.CreateV2(request)
	}
	return c.client.Create(request)
}

// UpdateIssue updates an existing issue.
func (c *JiraClient) UpdateIssue(key string, request *jira.EditRequest) error {
	// The jira package only has Edit method, no EditV2
	return c.client.Edit(key, request)
}

// DeleteIssue deletes an issue.
func (c *JiraClient) DeleteIssue(key string, cascade bool) error {
	return c.client.DeleteIssue(key, cascade)
}

// AssignIssue assigns an issue to a user.
func (c *JiraClient) AssignIssue(key string, assignee string) error {
	if c.installationType == jira.InstallationTypeLocal {
		return c.client.AssignIssueV2(key, assignee)
	}
	return c.client.AssignIssue(key, assignee)
}

// TransitionIssue transitions an issue to a new status.
func (c *JiraClient) TransitionIssue(key string, request *jira.TransitionRequest) error {
	_, err := c.client.Transition(key, request)
	return err
}

// AddComment adds a comment to an issue.
func (c *JiraClient) AddComment(key string, comment string, internal bool) error {
	return c.client.AddIssueComment(key, comment, internal)
}

// GetTransitions gets available transitions for an issue.
func (c *JiraClient) GetTransitions(key string) ([]*jira.Transition, error) {
	if c.installationType == jira.InstallationTypeLocal {
		return c.client.TransitionsV2(key)
	}
	return c.client.Transitions(key)
}

// GetProjects lists all accessible projects.
func (c *JiraClient) GetProjects() ([]*jira.Project, error) {
	return c.client.Project()
}

// GetProject gets a single project by key.
func (c *JiraClient) GetProject(key string) (*jira.Project, error) {
	// ProjectDetails doesn't exist, need to filter from all projects
	projects, err := c.client.Project()
	if err != nil {
		return nil, err
	}
	for _, project := range projects {
		if project.Key == key {
			return project, nil
		}
	}
	return nil, fmt.Errorf("project %s not found", key)
}

// GetBoards lists boards for a project.
func (c *JiraClient) GetBoards(project string, boardType string) (*jira.BoardResult, error) {
	return c.client.Boards(project, boardType)
}

// GetSprints lists sprints.
func (c *JiraClient) GetSprints(boardID int, state string, from, limit int) (*jira.SprintResult, error) {
	return c.client.Sprints(boardID, state, from, limit)
}

// GetSprintIssues lists issues in a sprint.
func (c *JiraClient) GetSprintIssues(sprintID int, jql string, from, limit uint) (*jira.SearchResult, error) {
	return c.client.SprintIssues(sprintID, jql, from, limit)
}

// GetEpics searches for epics using JQL.
// For board-specific epics, construct appropriate JQL query.
func (c *JiraClient) GetEpics(project string, from, limit uint) (*jira.SearchResult, error) {
	// Search for epics using JQL
	jql := fmt.Sprintf("project = %s AND issuetype = Epic", project)
	return c.SearchIssues(jql, from, limit)
}

// GetEpicIssues lists issues in an epic.
func (c *JiraClient) GetEpicIssues(epicKey, jql string, from, limit uint) (*jira.SearchResult, error) {
	return c.client.EpicIssues(epicKey, jql, from, limit)
}

// GetMyself gets information about the authenticated user.
func (c *JiraClient) GetMyself() (*jira.Me, error) {
	return c.client.Me()
}

// GetServerInfo gets server information.
func (c *JiraClient) GetServerInfo() (*jira.ServerInfo, error) {
	return c.client.ServerInfo()
}

// GetRawClient returns the underlying jira.Client for advanced usage.
// Use this when you need access to methods not exposed by JiraClient.
func (c *JiraClient) GetRawClient() *jira.Client {
	return c.client
}

// GetAllIssuesOptions contains options for fetching all issues.
type GetAllIssuesOptions struct {
	// Project filters by project key (optional)
	Project string
	
	// StartDate filters issues created or updated after this date (optional)
	// Format: "2006-01-02" or "2006-01-02 15:04"
	StartDate string
	
	// DateField specifies which date field to filter on: "created", "updated", or "resolved"
	// Default is "created"
	DateField string
	
	// MaxResults is the maximum number of issues to return (0 for no limit)
	MaxResults int
	
	// JQL allows passing custom JQL to combine with other filters
	JQL string
	
	// OrderBy specifies the field to order by (default: "created DESC")
	OrderBy string
}

// GetAllIssues fetches all issues with optional filtering.
// This method handles pagination automatically to retrieve all matching issues.
func (c *JiraClient) GetAllIssues(options GetAllIssuesOptions) ([]*jira.Issue, error) {
	// Build JQL query
	var jqlParts []string
	
	// Add project filter if specified
	if options.Project != "" {
		jqlParts = append(jqlParts, fmt.Sprintf("project = %s", options.Project))
	}
	
	// Add date filter if specified
	if options.StartDate != "" {
		dateField := options.DateField
		if dateField == "" {
			dateField = "created"
		}
		jqlParts = append(jqlParts, fmt.Sprintf("%s >= '%s'", dateField, options.StartDate))
	}
	
	// Add custom JQL if provided
	if options.JQL != "" {
		jqlParts = append(jqlParts, fmt.Sprintf("(%s)", options.JQL))
	}
	
	// Combine all JQL parts
	jql := ""
	if len(jqlParts) > 0 {
		jql = strings.Join(jqlParts, " AND ")
	}
	
	// Add ordering
	if options.OrderBy != "" {
		jql += fmt.Sprintf(" ORDER BY %s", options.OrderBy)
	} else {
		jql += " ORDER BY created DESC"
	}
	
	// Fetch all issues with pagination
	var allIssues []*jira.Issue
	const batchSize = 100
	var startAt uint = 0
	totalFetched := 0
	
	for {
		// Fetch a batch of issues
		results, err := c.SearchIssues(jql, startAt, batchSize)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch issues at offset %d: %w", startAt, err)
		}
		
		// Add issues to our collection
		allIssues = append(allIssues, results.Issues...)
		totalFetched += len(results.Issues)
		
		// Check if we've reached the limit (if set)
		if options.MaxResults > 0 && totalFetched >= options.MaxResults {
			// Trim to exact max results
			if len(allIssues) > options.MaxResults {
				allIssues = allIssues[:options.MaxResults]
			}
			break
		}
		
		// Check if we've fetched all issues
		if startAt+uint(len(results.Issues)) >= uint(results.Total) {
			break
		}
		
		// No more issues returned
		if len(results.Issues) == 0 {
			break
		}
		
		// Prepare for next batch
		startAt += batchSize
	}
	
	return allIssues, nil
}

// GetIssuesByDateRange fetches issues created or updated within a date range.
func (c *JiraClient) GetIssuesByDateRange(startDate, endDate string, dateField string) ([]*jira.Issue, error) {
	if dateField == "" {
		dateField = "created"
	}
	
	jql := fmt.Sprintf("%s >= '%s' AND %s <= '%s' ORDER BY %s DESC", 
		dateField, startDate, dateField, endDate, dateField)
	
	var allIssues []*jira.Issue
	const batchSize = 100
	var startAt uint = 0
	
	for {
		results, err := c.SearchIssues(jql, startAt, batchSize)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch issues: %w", err)
		}
		
		allIssues = append(allIssues, results.Issues...)
		
		if startAt+uint(len(results.Issues)) >= uint(results.Total) || len(results.Issues) == 0 {
			break
		}
		
		startAt += batchSize
	}
	
	return allIssues, nil
}

// GetRecentIssues fetches issues from the last N days.
func (c *JiraClient) GetRecentIssues(days int, project string) ([]*jira.Issue, error) {
	options := GetAllIssuesOptions{
		Project:   project,
		StartDate: fmt.Sprintf("-%dd", days),
		DateField: "created",
		OrderBy:   "created DESC",
	}
	return c.GetAllIssues(options)
}

// StatusChange represents a status transition in issue history.
type StatusChange struct {
	// Timestamp when the status change occurred
	Timestamp time.Time
	
	// Author who made the status change
	Author string
	
	// FromStatus is the previous status (empty for initial status)
	FromStatus string
	
	// ToStatus is the new status
	ToStatus string
	
	// AuthorEmail is the email of the person who made the change
	AuthorEmail string
	
	// AuthorDisplayName is the display name of the person who made the change
	AuthorDisplayName string
}

// IssueWithChangelog represents an issue with its changelog data.
type IssueWithChangelog struct {
	Key       string                 `json:"key"`
	Fields    jira.IssueFields       `json:"fields"`
	Changelog *IssueChangelog        `json:"changelog"`
}

// IssueChangelog represents the changelog section of an issue.
type IssueChangelog struct {
	StartAt    int                   `json:"startAt"`
	MaxResults int                   `json:"maxResults"`
	Total      int                   `json:"total"`
	Histories  []IssueHistory        `json:"histories"`
}

// IssueHistory represents a single history entry.
type IssueHistory struct {
	ID      string                `json:"id"`
	Author  *HistoryAuthor        `json:"author"`
	Created string                `json:"created"`
	Items   []HistoryItem         `json:"items"`
}

// HistoryAuthor represents the author of a history change.
type HistoryAuthor struct {
	Name         string `json:"name"`
	EmailAddress string `json:"emailAddress"`
	DisplayName  string `json:"displayName"`
}

// HistoryItem represents a single field change in history.
type HistoryItem struct {
	Field      string `json:"field"`
	FieldType  string `json:"fieldtype"`
	From       string `json:"from"`
	FromString string `json:"fromString"`
	To         string `json:"to"`
	ToString   string `json:"toString"`
}

// GetIssueStatusChanges retrieves all status changes for an issue.
// It fetches the issue with its changelog and extracts status transitions.
func (c *JiraClient) GetIssueStatusChanges(issueKey string) ([]StatusChange, error) {
	// Fetch issue with expanded changelog
	issueWithHistory, err := c.getIssueWithChangelog(issueKey)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issue with changelog: %w", err)
	}
	
	// Extract status changes from changelog
	var statusChanges []StatusChange
	
	if issueWithHistory.Changelog != nil {
		for _, history := range issueWithHistory.Changelog.Histories {
			for _, item := range history.Items {
				if item.Field == "status" {
					// Parse the timestamp
					timestamp, err := time.Parse(time.RFC3339, history.Created)
					if err != nil {
						// Try alternative format
						timestamp, err = time.Parse("2006-01-02T15:04:05.000-0700", history.Created)
						if err != nil {
							// Use current time as fallback
							timestamp = time.Now()
						}
					}
					
					change := StatusChange{
						Timestamp:  timestamp,
						FromStatus: item.FromString,
						ToStatus:   item.ToString,
					}
					
					// Add author information if available
					if history.Author != nil {
						change.Author = history.Author.Name
						change.AuthorEmail = history.Author.EmailAddress
						change.AuthorDisplayName = history.Author.DisplayName
					}
					
					statusChanges = append(statusChanges, change)
				}
			}
		}
		
		// Check if we need to fetch more history (pagination)
		if issueWithHistory.Changelog.Total > issueWithHistory.Changelog.StartAt+issueWithHistory.Changelog.MaxResults {
			// Fetch additional history pages
			additionalChanges, err := c.fetchAdditionalHistory(issueKey, issueWithHistory.Changelog.MaxResults)
			if err == nil {
				statusChanges = append(statusChanges, additionalChanges...)
			}
		}
	}
	
	// Add the initial status as the first change (if we have the created date)
	if len(statusChanges) > 0 && issueWithHistory.Fields.Created != "" {
		createdTime, err := time.Parse(time.RFC3339, issueWithHistory.Fields.Created)
		if err != nil {
			createdTime, err = time.Parse("2006-01-02T15:04:05.000-0700", issueWithHistory.Fields.Created)
		}
		if err == nil {
			// Find the earliest status change to determine the initial status
			earliestChange := statusChanges[len(statusChanges)-1]
			if earliestChange.FromStatus != "" {
				initialChange := StatusChange{
					Timestamp:  createdTime,
					FromStatus: "",
					ToStatus:   earliestChange.FromStatus,
				}
				if issueWithHistory.Fields.Reporter.Name != "" {
					initialChange.Author = issueWithHistory.Fields.Reporter.Name
					initialChange.AuthorDisplayName = issueWithHistory.Fields.Reporter.Name
				}
				statusChanges = append(statusChanges, initialChange)
			}
		}
	}
	
	// Sort by timestamp (oldest first)
	for i := 0; i < len(statusChanges)/2; i++ {
		j := len(statusChanges) - 1 - i
		statusChanges[i], statusChanges[j] = statusChanges[j], statusChanges[i]
	}
	
	return statusChanges, nil
}

// getIssueWithChangelog fetches an issue with its changelog expanded.
func (c *JiraClient) getIssueWithChangelog(issueKey string) (*IssueWithChangelog, error) {
	path := fmt.Sprintf("/issue/%s?expand=changelog", issueKey)
	
	ctx := context.Background()
	var httpRes *http.Response
	var err error
	
	if c.installationType == jira.InstallationTypeLocal {
		httpRes, err = c.client.GetV2(ctx, path, nil)
	} else {
		httpRes, err = c.client.Get(ctx, path, nil)
	}
	
	if err != nil {
		return nil, err
	}
	if httpRes == nil {
		return nil, fmt.Errorf("empty response")
	}
	defer httpRes.Body.Close()
	
	if httpRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", httpRes.StatusCode)
	}
	
	var issue IssueWithChangelog
	decoder := json.NewDecoder(httpRes.Body)
	if err := decoder.Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to decode issue with changelog: %w", err)
	}
	
	return &issue, nil
}

// fetchAdditionalHistory fetches additional history pages if changelog is paginated.
func (c *JiraClient) fetchAdditionalHistory(issueKey string, startAt int) ([]StatusChange, error) {
	var allChanges []StatusChange
	currentStart := startAt
	ctx := context.Background()
	
	for {
		path := fmt.Sprintf("/issue/%s/changelog?startAt=%d", issueKey, currentStart)
		
		var httpRes *http.Response
		var err error
		
		if c.installationType == jira.InstallationTypeLocal {
			httpRes, err = c.client.GetV2(ctx, path, nil)
		} else {
			httpRes, err = c.client.Get(ctx, path, nil)
		}
		
		if err != nil {
			return allChanges, err
		}
		if httpRes == nil {
			return allChanges, fmt.Errorf("empty response")
		}
		defer httpRes.Body.Close()
		
		if httpRes.StatusCode != http.StatusOK {
			return allChanges, fmt.Errorf("unexpected status code: %d", httpRes.StatusCode)
		}
		
		var changelog IssueChangelog
		decoder := json.NewDecoder(httpRes.Body)
		if err := decoder.Decode(&changelog); err != nil {
			return allChanges, err
		}
		
		// Extract status changes
		for _, history := range changelog.Histories {
			for _, item := range history.Items {
				if item.Field == "status" {
					timestamp, _ := time.Parse(time.RFC3339, history.Created)
					if timestamp.IsZero() {
						timestamp, _ = time.Parse("2006-01-02T15:04:05.000-0700", history.Created)
					}
					
					change := StatusChange{
						Timestamp:  timestamp,
						FromStatus: item.FromString,
						ToStatus:   item.ToString,
					}
					
					if history.Author != nil {
						change.Author = history.Author.Name
						change.AuthorEmail = history.Author.EmailAddress
						change.AuthorDisplayName = history.Author.DisplayName
					}
					
					allChanges = append(allChanges, change)
				}
			}
		}
		
		// Check if we need more pages
		if currentStart+len(changelog.Histories) >= changelog.Total {
			break
		}
		
		currentStart += len(changelog.Histories)
	}
	
	return allChanges, nil
}