package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	debug      bool
}

func New(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) SetDebug(debug bool) {
	c.debug = debug
}

func (c *Client) SetTimeout(d time.Duration) {
	c.httpClient.Timeout = d
}

func (c *Client) Get(ctx context.Context, path string, params url.Values, result any) error {
	return c.do(ctx, "GET", "/rest/api/1.0"+path, params, nil, result)
}

func (c *Client) Post(ctx context.Context, path string, body any, params url.Values, result any) error {
	return c.do(ctx, "POST", "/rest/api/1.0"+path, params, body, result)
}

func (c *Client) Put(ctx context.Context, path string, body any, params url.Values, result any) error {
	return c.do(ctx, "PUT", "/rest/api/1.0"+path, params, body, result)
}

func (c *Client) Delete(ctx context.Context, path string, params url.Values, result any) error {
	return c.do(ctx, "DELETE", "/rest/api/1.0"+path, params, nil, result)
}

func (c *Client) DeleteWithBody(ctx context.Context, path string, body any, params url.Values, result any) error {
	return c.do(ctx, "DELETE", "/rest/api/1.0"+path, params, body, result)
}

func (c *Client) DoAbsolute(ctx context.Context, method, path string, body any, params url.Values, result any) error {
	return c.do(ctx, method, path, params, body, result)
}

func (c *Client) GetRaw(ctx context.Context, path string, params url.Values) (string, error) {
	fullURL := c.baseURL + "/rest/api/1.0" + path
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

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
	c.debugLog("GET", path, resp.StatusCode, time.Since(start))

	if resp.StatusCode >= 400 {
		return "", c.parseError(resp)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *Client) Search(ctx context.Context, body any, params url.Values, result any) error {
	err := c.do(ctx, "POST", "/rest/search/latest/search", nil, body, result)
	if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == 405 {
		return c.do(ctx, "GET", "/rest/search/latest/search", params, nil, result)
	}
	return err
}

func (c *Client) do(ctx context.Context, method, path string, params url.Values, body any, result any) error {
	var lastErr error
	delays := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

	for attempt := 0; attempt <= len(delays); attempt++ {
		if attempt > 0 {
			time.Sleep(delays[attempt-1])
		}

		err := c.doOnce(ctx, method, path, params, body, result)
		if err == nil {
			return nil
		}

		apiErr, ok := err.(*APIError)
		if !ok || (apiErr.StatusCode != 429 && apiErr.StatusCode != 503) {
			return err
		}
		lastErr = err
		if attempt >= len(delays) {
			break
		}
	}
	return lastErr
}

func (c *Client) doOnce(ctx context.Context, method, path string, params url.Values, body any, result any) error {
	fullURL := c.baseURL + path
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return err
	}
	c.setHeaders(req)

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	c.debugLog(method, path, resp.StatusCode, time.Since(start))

	if resp.StatusCode >= 400 {
		return c.parseError(resp)
	}

	if resp.StatusCode == 204 || result == nil {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
}

func (c *Client) parseError(resp *http.Response) error {
	apiErr := &APIError{StatusCode: resp.StatusCode}

	if resp.StatusCode >= 500 {
		apiErr.Errors = []APIErrorDetail{{Message: fmt.Sprintf("Server error (%d)", resp.StatusCode)}}
		return apiErr
	}

	var errBody struct {
		Errors []APIErrorDetail `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&errBody); err == nil {
		apiErr.Errors = errBody.Errors
	}
	if len(apiErr.Errors) == 0 {
		apiErr.Errors = []APIErrorDetail{{Message: "Request failed"}}
	}
	return apiErr
}

func (c *Client) debugLog(method, path string, status int, duration time.Duration) {
	if c.debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] %s %s → %d (%s)\n", method, path, status, duration.Round(time.Millisecond))
	}
}
