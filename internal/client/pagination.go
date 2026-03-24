package client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/manu/bb/internal/validation"
)

// PagedResponse represents a paginated API response.
type PagedResponse[T any] struct {
	Values        []T  `json:"values"`
	Size          int  `json:"size"`
	Start         int  `json:"start"`
	Limit         int  `json:"limit"`
	IsLastPage    bool `json:"isLastPage"`
	NextPageStart int  `json:"nextPageStart"`
}

// GetPaged fetches a single page of results.
func GetPaged[T any](ctx context.Context, c *Client, path string, params url.Values, start, limit int) (*PagedResponse[T], error) {
	if params == nil {
		params = url.Values{}
	}
	params.Set("start", fmt.Sprintf("%d", validation.ClampStart(start)))
	params.Set("limit", fmt.Sprintf("%d", validation.ClampLimit(limit)))

	var result PagedResponse[T]
	if err := c.Get(ctx, path, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetAll fetches all pages by following nextPageStart until isLastPage is true.
func GetAll[T any](ctx context.Context, c *Client, path string, params url.Values, limit int) ([]T, error) {
	var all []T
	start := 0
	clampedLimit := validation.ClampLimit(limit)

	for {
		page, err := GetPaged[T](ctx, c, path, params, start, clampedLimit)
		if err != nil {
			return nil, err
		}
		all = append(all, page.Values...)
		if page.IsLastPage {
			break
		}
		start = page.NextPageStart
	}
	return all, nil
}
