package client

import (
	"context"
	"fmt"
	"net/url"
)

func (c *Client) ListCommits(ctx context.Context, project, repo string, until, since, path string, start, limit int) (*PagedResponse[Commit], error) {
	params := url.Values{}
	if until != "" {
		params.Set("until", until)
	}
	if since != "" {
		params.Set("since", since)
	}
	if path != "" {
		params.Set("path", path)
	}
	return GetPaged[Commit](ctx, c, repoPath(project, repo)+"/commits", params, start, limit)
}

func (c *Client) GetCommit(ctx context.Context, project, repo, commitID string) (*Commit, error) {
	var result Commit
	err := c.Get(ctx, fmt.Sprintf("%s/commits/%s", repoPath(project, repo), commitID), nil, &result)
	return &result, err
}

func (c *Client) GetCommitDiff(ctx context.Context, project, repo, commitID string, contextLines int, srcPath string) (*Diff, error) {
	params := url.Values{"contextLines": {fmt.Sprintf("%d", contextLines)}}
	if srcPath != "" {
		params.Set("srcPath", srcPath)
	}
	var result Diff
	err := c.Get(ctx, fmt.Sprintf("%s/commits/%s/diff", repoPath(project, repo), commitID), params, &result)
	return &result, err
}

func (c *Client) GetCommitChanges(ctx context.Context, project, repo, commitID string, start, limit int) (*PagedResponse[Change], error) {
	return GetPaged[Change](ctx, c, fmt.Sprintf("%s/commits/%s/changes", repoPath(project, repo), commitID), nil, start, limit)
}
