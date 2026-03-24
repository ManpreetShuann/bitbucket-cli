package client

import (
	"context"
)

func repoPath(project, repo string) string {
	return "/projects/" + project + "/repos/" + repo
}

func (c *Client) ListRepositories(ctx context.Context, project string, start, limit int) (*PagedResponse[Repository], error) {
	return GetPaged[Repository](ctx, c, "/projects/"+project+"/repos", nil, start, limit)
}

func (c *Client) GetRepository(ctx context.Context, project, repo string) (*Repository, error) {
	var result Repository
	err := c.Get(ctx, repoPath(project, repo), nil, &result)
	return &result, err
}

func (c *Client) CreateRepository(ctx context.Context, project, name, description string, forkable bool) (*Repository, error) {
	body := map[string]any{
		"name":     name,
		"scmId":    "git",
		"forkable": forkable,
	}
	if description != "" {
		body["description"] = description
	}
	var result Repository
	err := c.Post(ctx, "/projects/"+project+"/repos", body, nil, &result)
	return &result, err
}
