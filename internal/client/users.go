package client

import (
	"context"
	"net/url"
)

func (c *Client) FindUser(ctx context.Context, query string, start, limit int) (*PagedResponse[User], error) {
	params := url.Values{"filter": {query}}
	return GetPaged[User](ctx, c, "/users", params, start, limit)
}
