package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (c *Client) FindUser(ctx context.Context, query string, start, limit int) (*PagedResponse[User], error) {
	params := url.Values{"filter": {query}}
	return GetPaged[User](ctx, c, "/users", params, start, limit)
}

// GetUser fetches a single user by their slug (username).
func (c *Client) GetUser(ctx context.Context, slug string) (*User, error) {
	var user User
	if err := c.Get(ctx, "/users/"+slug, nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// CurrentUser returns the authenticated user by reading the X-AUSERNAME response
// header that Bitbucket Server sets on every authenticated request.
func (c *Client) CurrentUser(ctx context.Context) (*User, error) {
	slug, err := c.currentUsername(ctx)
	if err != nil {
		return nil, err
	}
	return c.GetUser(ctx, slug)
}

// currentUsername makes a lightweight request and reads the X-AUSERNAME header.
func (c *Client) currentUsername(ctx context.Context) (string, error) {
	fullURL := c.baseURL + "/rest/api/1.0/application-properties"

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return "", err
	}
	c.setHeaders(req)

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	c.debugLog("GET", "/rest/api/1.0/application-properties", resp.StatusCode, time.Since(start))

	if resp.StatusCode >= 400 {
		return "", c.parseError(resp)
	}

	username := strings.TrimSpace(resp.Header.Get("X-AUSERNAME"))
	if username == "" {
		return "", fmt.Errorf("could not determine authenticated user: X-AUSERNAME header missing")
	}
	return username, nil
}
