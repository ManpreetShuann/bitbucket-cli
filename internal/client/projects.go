package client

import (
	"context"
)

func (c *Client) ListProjects(ctx context.Context, start, limit int) (*PagedResponse[Project], error) {
	return GetPaged[Project](ctx, c, "/projects", nil, start, limit)
}

func (c *Client) GetProject(ctx context.Context, key string) (*Project, error) {
	var result Project
	err := c.Get(ctx, "/projects/"+key, nil, &result)
	return &result, err
}
