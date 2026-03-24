package client

import (
	"context"
	"fmt"
)

func (c *Client) GetAttachment(ctx context.Context, project, repo string, attachmentID string) (string, error) {
	return c.GetRaw(ctx, fmt.Sprintf("%s/attachments/%s", repoPath(project, repo), attachmentID), nil)
}

func (c *Client) GetAttachmentMetadata(ctx context.Context, project, repo string, attachmentID string) (map[string]any, error) {
	var result map[string]any
	err := c.Get(ctx, fmt.Sprintf("%s/attachments/%s/metadata", repoPath(project, repo), attachmentID), nil, &result)
	return result, err
}

func (c *Client) SaveAttachmentMetadata(ctx context.Context, project, repo string, attachmentID string, metadata map[string]any) (map[string]any, error) {
	var result map[string]any
	err := c.Put(ctx, fmt.Sprintf("%s/attachments/%s/metadata", repoPath(project, repo), attachmentID), metadata, nil, &result)
	return result, err
}
