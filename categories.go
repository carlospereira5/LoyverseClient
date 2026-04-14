package loyverse

import (
	"context"
	"net/url"
)

// ListCategories returns all product categories, automatically following pagination cursors.
func (c *Client) ListCategories(ctx context.Context) ([]Category, error) {
	return paginate(func(cursor string) ([]Category, string, error) {
		params := url.Values{"limit": {_pageLimit}}
		if cursor != "" {
			params.Set("cursor", cursor)
		}
		var resp categoriesResponse
		if err := c.get(ctx, "/categories", params, &resp); err != nil {
			return nil, "", err
		}
		return resp.Categories, resp.Cursor, nil
	})
}
