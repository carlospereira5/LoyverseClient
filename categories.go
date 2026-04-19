package loyverse

import (
	"context"
	"fmt"
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

// CreateOrUpdateCategory creates a new category or updates an existing one via POST /categories.
// Set req.ID to update; omit it to create.
func (c *Client) CreateOrUpdateCategory(ctx context.Context, req CategoryRequest) (*Category, error) {
	var cat Category
	if err := c.post(ctx, "/categories", req, &cat); err != nil {
		return nil, fmt.Errorf("loyverse: create or update category: %w", err)
	}
	return &cat, nil
}

// DeleteCategory permanently deletes the category with the given ID via DELETE /categories/:id.
func (c *Client) DeleteCategory(ctx context.Context, id string) error {
	if err := c.delete(ctx, "/categories/"+id); err != nil {
		return fmt.Errorf("loyverse: delete category %s: %w", id, err)
	}
	return nil
}
