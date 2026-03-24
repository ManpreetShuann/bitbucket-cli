package client

import (
	"context"
	"net/url"
)

func (c *Client) BrowseFiles(ctx context.Context, project, repo, path, at string, start, limit int) (*PagedResponse[FileEntry], error) {
	params := url.Values{}
	if at != "" {
		params.Set("at", at)
	}
	apiPath := repoPath(project, repo) + "/browse"
	if path != "" {
		apiPath += "/" + path
	}
	return GetPaged[FileEntry](ctx, c, apiPath, params, start, limit)
}

func (c *Client) GetFileContent(ctx context.Context, project, repo, path, at string) (string, error) {
	params := url.Values{}
	if at != "" {
		params.Set("at", at)
	}
	return c.GetRaw(ctx, repoPath(project, repo)+"/raw/"+path, params)
}

func (c *Client) ListFiles(ctx context.Context, project, repo, path, at string, start, limit int) (*PagedResponse[string], error) {
	params := url.Values{}
	if at != "" {
		params.Set("at", at)
	}
	if path != "" {
		params.Set("path", path)
	}
	return GetPaged[string](ctx, c, repoPath(project, repo)+"/files", params, start, limit)
}
