package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// SearchResult struct holds response from /search endpoint.
type SearchResult struct {
	StartAt       int      `json:"startAt"`
	MaxResults    int      `json:"maxResults"`
	Total         int      `json:"total"`
	Issues        []*Issue `json:"issues"`
	NextPageToken string   `json:"nextPageToken,omitempty"` // New field for cloud pagination
	IsLast        bool     `json:"isLast,omitempty"`        // New field to indicate last page
}

// Search searches for issues using v3 version of the Jira GET /search endpoint.
func (c *Client) Search(jql string, from, limit uint) (*SearchResult, error) {
	return c.search(jql, from, limit, apiVersion3)
}

// SearchV2 searches an issues using v2 version of the Jira GET /search endpoint.
func (c *Client) SearchV2(jql string, from, limit uint) (*SearchResult, error) {
	return c.search(jql, from, limit, apiVersion2)
}

func (c *Client) search(jql string, from, limit uint, ver string) (*SearchResult, error) {
	var (
		res *http.Response
		err error
	)

	// For cloud instances, use the new /search/jql endpoint
	// The new endpoint requires bounded queries, so we add a default bound if missing
	if ver == apiVersion3 {
		// Check if JQL is bounded (has restrictions like created >= -Xd, project = X, etc.)
		// If not bounded, add a default restriction to avoid "Unbounded JQL queries are not allowed" error
		if !isJQLBounded(jql) {
			// Add a default bound - issues created in last 90 days
			if jql == "" {
				jql = "created >= -90d"
			} else {
				jql = fmt.Sprintf("created >= -90d AND (%s)", jql)
			}
		}
		
		// Use the new search/jql endpoint with fields=*all to get all fields
		path := fmt.Sprintf("/search/jql?jql=%s&startAt=%d&maxResults=%d&fields=*all", 
			url.QueryEscape(jql), from, limit)
		res, err = c.Get(context.Background(), path, nil)
	} else {
		// For v2 (server/datacenter), use the old endpoint
		path := fmt.Sprintf("/search?jql=%s&startAt=%d&maxResults=%d", 
			url.QueryEscape(jql), from, limit)
		res, err = c.GetV2(context.Background(), path, nil)
	}

	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, ErrEmptyResponse
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, formatUnexpectedResponse(res)
	}

	var out SearchResult
	err = json.NewDecoder(res.Body).Decode(&out)
	if err != nil {
		return nil, err
	}

	// For the new endpoint, Total might not be provided, calculate it from response
	if ver == apiVersion3 && out.Total == 0 && len(out.Issues) > 0 {
		// If IsLast is true, total is startAt + number of issues
		if out.IsLast {
			out.Total = int(from) + len(out.Issues)
		} else {
			// Otherwise, we don't know the exact total, set to a large number
			out.Total = 10000 // This is a reasonable upper bound
		}
	}

	return &out, err
}

// isJQLBounded checks if a JQL query has sufficient restrictions for the new API
func isJQLBounded(jql string) bool {
	// Check for common bounding conditions
	boundingTerms := []string{
		"created >=", "created >", "created =", "created <=", "created <",
		"updated >=", "updated >", "updated =", "updated <=", "updated <",
		"project =", "project in", "project IN",
		"id =", "id in", "id IN",
		"key =", "key in", "key IN",
		"issuekey =", "issuekey in", "issuekey IN",
	}
	
	jqlLower := strings.ToLower(jql)
	for _, term := range boundingTerms {
		if strings.Contains(jqlLower, strings.ToLower(term)) {
			return true
		}
	}
	
	return false
}
