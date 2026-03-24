package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

func prPath(project, repo string, id int) string {
	return fmt.Sprintf("%s/pull-requests/%d", repoPath(project, repo), id)
}

// ListPROptions holds filter options for listing pull requests.
type ListPROptions struct {
	State       string
	Direction   string
	Order       string
	At          string
	FilterText  string
	Participant string
	Draft       *bool
	Start       int
	Limit       int
}

func (c *Client) ListPullRequests(ctx context.Context, project, repo string, opts ListPROptions) (*PagedResponse[PullRequest], error) {
	params := url.Values{}
	if opts.State != "" {
		params.Set("state", strings.ToUpper(opts.State))
	}
	if opts.Direction != "" {
		params.Set("direction", strings.ToUpper(opts.Direction))
	}
	if opts.Order != "" {
		params.Set("order", strings.ToUpper(opts.Order))
	}
	if opts.At != "" {
		params.Set("at", opts.At)
	}
	if opts.FilterText != "" {
		params.Set("filterText", opts.FilterText)
	}
	if opts.Participant != "" {
		params.Set("role.1", "PARTICIPANT")
		params.Set("username.1", opts.Participant)
	}
	if opts.Draft != nil {
		params.Set("draft", fmt.Sprintf("%t", *opts.Draft))
	}
	return GetPaged[PullRequest](ctx, c, repoPath(project, repo)+"/pull-requests", params, opts.Start, opts.Limit)
}

func (c *Client) GetPullRequest(ctx context.Context, project, repo string, id int) (*PullRequest, error) {
	var result PullRequest
	err := c.Get(ctx, prPath(project, repo, id), nil, &result)
	return &result, err
}

// CreatePRInput holds input for creating a pull request.
type CreatePRInput struct {
	Title       string
	Description string
	FromRef     string
	ToRef       string
	Reviewers   []string
	Draft       bool
}

func (c *Client) CreatePullRequest(ctx context.Context, project, repo string, input CreatePRInput) (*PullRequest, error) {
	fromRef := input.FromRef
	if !strings.HasPrefix(fromRef, "refs/") {
		fromRef = "refs/heads/" + fromRef
	}
	toRef := input.ToRef
	if !strings.HasPrefix(toRef, "refs/") {
		toRef = "refs/heads/" + toRef
	}

	body := map[string]any{
		"title":       input.Title,
		"description": input.Description,
		"fromRef": map[string]any{
			"id":         fromRef,
			"repository": map[string]any{"slug": repo, "project": map[string]any{"key": project}},
		},
		"toRef": map[string]any{
			"id":         toRef,
			"repository": map[string]any{"slug": repo, "project": map[string]any{"key": project}},
		},
	}
	if len(input.Reviewers) > 0 {
		reviewers := make([]map[string]any, len(input.Reviewers))
		for i, r := range input.Reviewers {
			reviewers[i] = map[string]any{"user": map[string]any{"name": r}}
		}
		body["reviewers"] = reviewers
	}
	if input.Draft {
		body["draft"] = true
	}

	var result PullRequest
	err := c.Post(ctx, repoPath(project, repo)+"/pull-requests", body, nil, &result)
	return &result, err
}

// UpdatePRInput holds input for updating a pull request.
type UpdatePRInput struct {
	Title        string
	Description  *string
	Reviewers    []string
	TargetBranch string
	Draft        *bool
}

func (c *Client) UpdatePullRequest(ctx context.Context, project, repo string, id, version int, input UpdatePRInput) (*PullRequest, error) {
	current, err := c.GetPullRequest(ctx, project, repo, id)
	if err != nil {
		return nil, err
	}

	body := map[string]any{
		"version":     version,
		"title":       current.Title,
		"description": current.Description,
		"toRef":       map[string]any{"id": current.ToRef.ID, "repository": map[string]any{"slug": repo, "project": map[string]any{"key": project}}},
		"reviewers":   current.Reviewers,
	}
	if input.Title != "" {
		body["title"] = input.Title
	}
	if input.Description != nil {
		body["description"] = *input.Description
	}
	if input.Reviewers != nil {
		reviewers := make([]map[string]any, len(input.Reviewers))
		for i, r := range input.Reviewers {
			reviewers[i] = map[string]any{"user": map[string]any{"name": r}}
		}
		body["reviewers"] = reviewers
	}
	if input.TargetBranch != "" {
		toRef := input.TargetBranch
		if !strings.HasPrefix(toRef, "refs/") {
			toRef = "refs/heads/" + toRef
		}
		body["toRef"] = map[string]any{
			"id":         toRef,
			"repository": map[string]any{"slug": repo, "project": map[string]any{"key": project}},
		}
	}
	if input.Draft != nil {
		body["draft"] = *input.Draft
	}

	var result PullRequest
	err = c.Put(ctx, prPath(project, repo, id), body, nil, &result)
	return &result, err
}

func (c *Client) MergePullRequest(ctx context.Context, project, repo string, id, version int, strategy string) (*PullRequest, error) {
	params := url.Values{"version": {fmt.Sprintf("%d", version)}}
	if strategy != "" {
		params.Set("strategyId", strategy)
	}
	var result PullRequest
	err := c.Post(ctx, prPath(project, repo, id)+"/merge", nil, params, &result)
	return &result, err
}

func (c *Client) DeclinePullRequest(ctx context.Context, project, repo string, id, version int) (*PullRequest, error) {
	params := url.Values{"version": {fmt.Sprintf("%d", version)}}
	var result PullRequest
	err := c.Post(ctx, prPath(project, repo, id)+"/decline", nil, params, &result)
	return &result, err
}

func (c *Client) ReopenPullRequest(ctx context.Context, project, repo string, id, version int) (*PullRequest, error) {
	params := url.Values{"version": {fmt.Sprintf("%d", version)}}
	var result PullRequest
	err := c.Post(ctx, prPath(project, repo, id)+"/reopen", nil, params, &result)
	return &result, err
}

func (c *Client) ApprovePullRequest(ctx context.Context, project, repo string, id int) (*Participant, error) {
	var result Participant
	err := c.Post(ctx, prPath(project, repo, id)+"/approve", nil, nil, &result)
	return &result, err
}

func (c *Client) UnapprovePullRequest(ctx context.Context, project, repo string, id int) error {
	return c.Delete(ctx, prPath(project, repo, id)+"/approve", nil, nil)
}

func (c *Client) RequestChanges(ctx context.Context, project, repo string, id int) (*Participant, error) {
	var result Participant
	err := c.Put(ctx, prPath(project, repo, id)+"/participants/status", map[string]any{"status": "NEEDS_WORK"}, nil, &result)
	return &result, err
}

func (c *Client) RemoveChangeRequest(ctx context.Context, project, repo string, id int) (*Participant, error) {
	var result Participant
	err := c.Put(ctx, prPath(project, repo, id)+"/participants/status", map[string]any{"status": "UNAPPROVED"}, nil, &result)
	return &result, err
}

func (c *Client) CanMerge(ctx context.Context, project, repo string, id int) (*MergeStatus, error) {
	var result MergeStatus
	err := c.Get(ctx, prPath(project, repo, id)+"/merge", nil, &result)
	return &result, err
}

func (c *Client) WatchPullRequest(ctx context.Context, project, repo string, id int) error {
	return c.Post(ctx, prPath(project, repo, id)+"/watch", nil, nil, nil)
}

func (c *Client) UnwatchPullRequest(ctx context.Context, project, repo string, id int) error {
	return c.Delete(ctx, prPath(project, repo, id)+"/watch", nil, nil)
}

func (c *Client) GetCommitMessageSuggestion(ctx context.Context, project, repo string, id int) (map[string]any, error) {
	var result map[string]any
	err := c.Get(ctx, prPath(project, repo, id)+"/commit-message-suggestion", nil, &result)
	return result, err
}

func (c *Client) GetPullRequestDiff(ctx context.Context, project, repo string, id, contextLines int, srcPath string) (*Diff, error) {
	params := url.Values{"contextLines": {fmt.Sprintf("%d", contextLines)}}
	if srcPath != "" {
		params.Set("srcPath", srcPath)
	}
	var result Diff
	err := c.Get(ctx, prPath(project, repo, id)+"/diff", params, &result)
	return &result, err
}

func (c *Client) GetPullRequestDiffStat(ctx context.Context, project, repo string, id, start, limit int) (*PagedResponse[Change], error) {
	return GetPaged[Change](ctx, c, prPath(project, repo, id)+"/changes", nil, start, limit)
}

func (c *Client) ListPullRequestCommits(ctx context.Context, project, repo string, id, start, limit int) (*PagedResponse[Commit], error) {
	return GetPaged[Commit](ctx, c, prPath(project, repo, id)+"/commits", nil, start, limit)
}

func (c *Client) GetPullRequestActivities(ctx context.Context, project, repo string, id, start, limit int) (*PagedResponse[Activity], error) {
	return GetPaged[Activity](ctx, c, prPath(project, repo, id)+"/activities", nil, start, limit)
}

func (c *Client) ListPullRequestParticipants(ctx context.Context, project, repo string, id, start, limit int) (*PagedResponse[Participant], error) {
	return GetPaged[Participant](ctx, c, prPath(project, repo, id)+"/participants", nil, start, limit)
}

func (c *Client) PublishDraft(ctx context.Context, project, repo string, id, version int) (*PullRequest, error) {
	current, err := c.GetPullRequest(ctx, project, repo, id)
	if err != nil {
		return nil, err
	}
	body := map[string]any{
		"version":     version,
		"title":       current.Title,
		"description": current.Description,
		"draft":       false,
		"toRef":       map[string]any{"id": current.ToRef.ID, "repository": map[string]any{"slug": repo, "project": map[string]any{"key": project}}},
		"reviewers":   current.Reviewers,
	}
	var result PullRequest
	err = c.Put(ctx, prPath(project, repo, id), body, nil, &result)
	return &result, err
}

func (c *Client) ConvertToDraft(ctx context.Context, project, repo string, id, version int) (*PullRequest, error) {
	current, err := c.GetPullRequest(ctx, project, repo, id)
	if err != nil {
		return nil, err
	}
	body := map[string]any{
		"version":     version,
		"title":       current.Title,
		"description": current.Description,
		"draft":       true,
		"toRef":       map[string]any{"id": current.ToRef.ID, "repository": map[string]any{"slug": repo, "project": map[string]any{"key": project}}},
		"reviewers":   current.Reviewers,
	}
	var result PullRequest
	err = c.Put(ctx, prPath(project, repo, id), body, nil, &result)
	return &result, err
}
