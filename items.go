package loyverse

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
)

// GetItems returns all items in the catalog, automatically following pagination cursors.
func (c *Client) GetItems(ctx context.Context) ([]Item, error) {
	return paginate(func(cursor string) ([]Item, string, error) {
		params := url.Values{"limit": {_pageLimit}}
		if cursor != "" {
			params.Set("cursor", cursor)
		}
		var resp itemsResponse
		if err := c.get(ctx, "/items", params, &resp); err != nil {
			return nil, "", err
		}
		return resp.Items, resp.Cursor, nil
	})
}

// GetItem returns a single item by its Loyverse ID.
func (c *Client) GetItem(ctx context.Context, id string) (*Item, error) {
	var item Item
	if err := c.get(ctx, "/items/"+id, nil, &item); err != nil {
		return nil, fmt.Errorf("loyverse: get item %s: %w", id, err)
	}
	return &item, nil
}

// CreateItem creates a new item via POST /items.
func (c *Client) CreateItem(ctx context.Context, req CreateItemRequest) (*Item, error) {
	var item Item
	if err := c.post(ctx, "/items", req, &item); err != nil {
		return nil, fmt.Errorf("loyverse: create item: %w", err)
	}
	return &item, nil
}

// SetItemCost updates the cost field of an item and all its variants.
// It fetches the current item as a raw map, modifies only the cost fields, and re-uploads
// the full object. This preserves all other fields (name, variants, stores, etc.) unchanged.
func (c *Client) SetItemCost(ctx context.Context, itemID string, cost float64) error {
	// Fetch as map[string]any to avoid accidentally zeroing fields our typed struct
	// does not capture (e.g., store-specific prices, modifier links).
	var raw map[string]any
	if err := c.get(ctx, "/items/"+itemID, nil, &raw); err != nil {
		return fmt.Errorf("loyverse: set item cost %s: fetch item: %w", itemID, err)
	}

	raw["track_stock"] = true
	raw["cost"] = cost
	if variants, ok := raw["variants"].([]any); ok {
		for _, v := range variants {
			if variant, ok := v.(map[string]any); ok {
				variant["cost"] = cost
				variant["purchase_cost"] = cost
			}
		}
	}

	if err := c.post(ctx, "/items", raw, nil); err != nil {
		return fmt.Errorf("loyverse: set item cost %s: update item: %w", itemID, err)
	}
	return nil
}

// ResetAllCosts sets the cost of every item and all its variants to zero.
// It runs up to c.workers concurrent requests.
// Returns (successful, failed) counts; per-item errors are logged but do not abort the operation.
func (c *Client) ResetAllCosts(ctx context.Context) (ok, failed int, err error) {
	items, err := c.GetItems(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("loyverse: reset all costs: %w", err)
	}

	jobs := make(chan Item, len(items))
	for _, item := range items {
		jobs <- item
	}
	close(jobs)

	var okCount, failCount int64
	var wg sync.WaitGroup

	for range c.workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range jobs {
				if setErr := c.SetItemCost(ctx, item.ID, 0); setErr != nil {
					c.logger.ErrorContext(ctx, "loyverse: reset cost failed",
						"item_id", item.ID,
						"item_name", item.Name,
						"err", setErr,
					)
					atomic.AddInt64(&failCount, 1)
					continue
				}
				atomic.AddInt64(&okCount, 1)
			}
		}()
	}
	wg.Wait()

	return int(okCount), int(failCount), nil
}
