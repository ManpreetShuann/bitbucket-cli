package client

import (
	"context"
	"net/url"
)

// SearchCodeResult wraps the search API response.
type SearchCodeResult struct {
	Results []SearchResult `json:"values"`
	Count   int            `json:"count"`
}

func (c *Client) SearchCode(ctx context.Context, query, project, repo string, start, limit int) (*SearchCodeResult, error) {
	body := map[string]any{
		"query": query,
		"type":  "CODE",
	}
	if project != "" {
		body["projectKey"] = project
	}
	if repo != "" {
		body["repositorySlug"] = repo
	}
	body["start"] = start
	body["limit"] = limit

	params := url.Values{
		"query": {query},
		"type":  {"CODE"},
	}
	if project != "" {
		params.Set("projectKey", project)
	}
	if repo != "" {
		params.Set("repositorySlug", repo)
	}

	var result SearchCodeResult
	err := c.Search(ctx, body, params, &result)
	return &result, err
}

func (c *Client) FindFile(ctx context.Context, project, repo, pattern string, start, limit int) (*PagedResponse[string], error) {
	params := url.Values{}
	if pattern != "" {
		params.Set("filter", pattern)
	}
	return GetPaged[string](ctx, c, repoPath(project, repo)+"/files", params, start, limit)
}
