package client

import (
	"context"
	"fmt"
)

func (c *Client) ListPRTasks(ctx context.Context, project, repo string, prID, start, limit int) (*PagedResponse[Task], error) {
	return GetPaged[Task](ctx, c, prPath(project, repo, prID)+"/tasks", nil, start, limit)
}

func (c *Client) GetPRTask(ctx context.Context, project, repo string, prID, taskID int) (*Task, error) {
	var result Task
	err := c.Get(ctx, fmt.Sprintf("%s/tasks/%d", prPath(project, repo, prID), taskID), nil, &result)
	return &result, err
}

func (c *Client) CreatePRTask(ctx context.Context, project, repo string, prID int, text string, commentID *int) (*Task, error) {
	body := map[string]any{"text": text}
	if commentID != nil {
		body["anchor"] = map[string]any{"id": *commentID, "type": "COMMENT"}
	}
	var result Task
	err := c.Post(ctx, prPath(project, repo, prID)+"/tasks", body, nil, &result)
	return &result, err
}

func (c *Client) UpdatePRTask(ctx context.Context, project, repo string, prID, taskID int, text, state string) (*Task, error) {
	body := map[string]any{}
	if text != "" {
		body["text"] = text
	}
	if state != "" {
		body["state"] = state
	}
	var result Task
	err := c.Put(ctx, fmt.Sprintf("%s/tasks/%d", prPath(project, repo, prID), taskID), body, nil, &result)
	return &result, err
}

func (c *Client) DeletePRTask(ctx context.Context, project, repo string, prID, taskID int) error {
	return c.Delete(ctx, fmt.Sprintf("%s/tasks/%d", prPath(project, repo, prID), taskID), nil, nil)
}
