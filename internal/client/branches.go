package client

import (
	"context"
	"net/url"
)

func (c *Client) ListBranches(ctx context.Context, project, repo, filter string, start, limit int) (*PagedResponse[Branch], error) {
	params := url.Values{}
	if filter != "" {
		params.Set("filterText", filter)
	}
	return GetPaged[Branch](ctx, c, repoPath(project, repo)+"/branches", params, start, limit)
}

func (c *Client) GetDefaultBranch(ctx context.Context, project, repo string) (*Branch, error) {
	var result Branch
	err := c.Get(ctx, repoPath(project, repo)+"/branches/default", nil, &result)
	return &result, err
}

func (c *Client) CreateBranch(ctx context.Context, project, repo, name, startPoint string) (*Branch, error) {
	body := map[string]any{"name": name, "startPoint": startPoint}
	var result Branch
	err := c.Post(ctx, repoPath(project, repo)+"/branches", body, nil, &result)
	return &result, err
}

func (c *Client) ListTags(ctx context.Context, project, repo, filter string, start, limit int) (*PagedResponse[Tag], error) {
	params := url.Values{}
	if filter != "" {
		params.Set("filterText", filter)
	}
	return GetPaged[Tag](ctx, c, repoPath(project, repo)+"/tags", params, start, limit)
}
