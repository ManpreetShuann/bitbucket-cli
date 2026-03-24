package client

import (
	"context"
	"fmt"
	"net/url"
)

func (c *Client) ListPRComments(ctx context.Context, project, repo string, prID, start, limit int) (*PagedResponse[Comment], error) {
	return GetPaged[Comment](ctx, c, prPath(project, repo, prID)+"/comments", nil, start, limit)
}

func (c *Client) GetPRComment(ctx context.Context, project, repo string, prID, commentID int) (*Comment, error) {
	var result Comment
	err := c.Get(ctx, fmt.Sprintf("%s/comments/%d", prPath(project, repo, prID), commentID), nil, &result)
	return &result, err
}

// AddCommentInput holds input for adding a PR comment.
type AddCommentInput struct {
	Text     string
	Severity string
	ParentID *int
	FilePath string
	Line     *int
	LineType string
	FileType string
}

func (c *Client) AddPRComment(ctx context.Context, project, repo string, prID int, input AddCommentInput) (*Comment, error) {
	body := map[string]any{"text": input.Text}
	if input.Severity != "" {
		body["severity"] = input.Severity
	}
	if input.ParentID != nil {
		body["parent"] = map[string]any{"id": *input.ParentID}
	}
	if input.FilePath != "" {
		anchor := map[string]any{"path": input.FilePath}
		if input.Line != nil {
			anchor["line"] = *input.Line
		}
		if input.LineType != "" {
			anchor["lineType"] = input.LineType
		}
		if input.FileType != "" {
			anchor["fileType"] = input.FileType
		}
		body["anchor"] = anchor
	}
	var result Comment
	err := c.Post(ctx, prPath(project, repo, prID)+"/comments", body, nil, &result)
	return &result, err
}

func (c *Client) UpdatePRComment(ctx context.Context, project, repo string, prID, commentID, version int, text string) (*Comment, error) {
	var result Comment
	err := c.Put(ctx, fmt.Sprintf("%s/comments/%d", prPath(project, repo, prID), commentID), map[string]any{"text": text, "version": version}, nil, &result)
	return &result, err
}

func (c *Client) ResolvePRComment(ctx context.Context, project, repo string, prID, commentID, version int) (*Comment, error) {
	var result Comment
	err := c.Put(ctx, fmt.Sprintf("%s/comments/%d", prPath(project, repo, prID), commentID), map[string]any{"state": "RESOLVED", "version": version}, nil, &result)
	return &result, err
}

func (c *Client) ReopenPRComment(ctx context.Context, project, repo string, prID, commentID, version int) (*Comment, error) {
	var result Comment
	err := c.Put(ctx, fmt.Sprintf("%s/comments/%d", prPath(project, repo, prID), commentID), map[string]any{"state": "OPEN", "version": version}, nil, &result)
	return &result, err
}

func (c *Client) DeletePRComment(ctx context.Context, project, repo string, prID, commentID, version int) error {
	params := url.Values{"version": {fmt.Sprintf("%d", version)}}
	return c.Delete(ctx, fmt.Sprintf("%s/comments/%d", prPath(project, repo, prID), commentID), params, nil)
}
