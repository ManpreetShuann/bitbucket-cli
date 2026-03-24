package client

import (
	"context"
	"fmt"
)

// DeleteBranch deletes a branch via the branch-utils API.
func (c *Client) DeleteBranch(ctx context.Context, project, repo, branchName string) error {
	body := map[string]any{"name": "refs/heads/" + branchName}
	return c.DoAbsolute(ctx, "DELETE", fmt.Sprintf("/rest/branch-utils/1.0%s/branches", repoPath(project, repo)), body, nil, nil)
}

// DeleteTag deletes a tag via the git API.
func (c *Client) DeleteTag(ctx context.Context, project, repo, tagName string) error {
	return c.DoAbsolute(ctx, "DELETE", fmt.Sprintf("/rest/git/1.0%s/tags/%s", repoPath(project, repo), tagName), nil, nil, nil)
}

// DeletePullRequest deletes a PR permanently.
func (c *Client) DeletePullRequest(ctx context.Context, project, repo string, id, version int) error {
	body := map[string]any{"version": version}
	return c.DeleteWithBody(ctx, prPath(project, repo, id), body, nil, nil)
}

// DeleteProject deletes a project permanently.
func (c *Client) DeleteProject(ctx context.Context, key string) error {
	return c.Delete(ctx, "/projects/"+key, nil, nil)
}

// DeleteRepository deletes a repository permanently.
func (c *Client) DeleteRepository(ctx context.Context, project, repo string) error {
	return c.Delete(ctx, repoPath(project, repo), nil, nil)
}

// DeleteAttachment deletes an attachment.
func (c *Client) DeleteAttachment(ctx context.Context, project, repo, attachmentID string) error {
	return c.Delete(ctx, fmt.Sprintf("%s/attachments/%s", repoPath(project, repo), attachmentID), nil, nil)
}

// DeleteAttachmentMetadata deletes attachment metadata.
func (c *Client) DeleteAttachmentMetadata(ctx context.Context, project, repo, attachmentID string) error {
	return c.Delete(ctx, fmt.Sprintf("%s/attachments/%s/metadata", repoPath(project, repo), attachmentID), nil, nil)
}
