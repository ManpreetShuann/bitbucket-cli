package client

import (
	"context"
	"net/url"
	"strings"
)

func (c *Client) ListDashboardPRs(ctx context.Context, state, role, order string, start, limit int) (*PagedResponse[PullRequest], error) {
	params := url.Values{}
	if state != "" {
		params.Set("state", strings.ToUpper(state))
	}
	if role != "" {
		params.Set("role", strings.ToUpper(role))
	}
	if order != "" {
		params.Set("order", strings.ToUpper(order))
	}
	return GetPaged[PullRequest](ctx, c, "/dashboard/pull-requests", params, start, limit)
}

func (c *Client) ListInboxPRs(ctx context.Context, start, limit int) (*PagedResponse[PullRequest], error) {
	return GetPaged[PullRequest](ctx, c, "/inbox/pull-requests", nil, start, limit)
}
