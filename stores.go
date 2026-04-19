package loyverse

import (
	"context"
	"fmt"
)

// ListStores returns all stores for the merchant account.
// The Loyverse /stores endpoint is not cursor-paginated; all results are returned in a single response.
func (c *Client) ListStores(ctx context.Context) ([]Store, error) {
	var resp storesResponse
	if err := c.get(ctx, "/stores", nil, &resp); err != nil {
		return nil, fmt.Errorf("loyverse: list stores: %w", err)
	}
	return resp.Stores, nil
}

// GetStore returns a single store by its Loyverse ID.
func (c *Client) GetStore(ctx context.Context, id string) (*Store, error) {
	var store Store
	if err := c.get(ctx, "/stores/"+id, nil, &store); err != nil {
		return nil, fmt.Errorf("loyverse: get store %s: %w", id, err)
	}
	return &store, nil
}
