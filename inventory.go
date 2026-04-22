package loyverse

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
)

// GetInventoryLevels returns all stock levels across all variants and stores,
// automatically following pagination cursors.
func (c *Client) GetInventoryLevels(ctx context.Context) ([]InventoryLevel, error) {
	return paginate(func(cursor string) ([]InventoryLevel, string, error) {
		params := url.Values{"limit": {_pageLimit}}
		if cursor != "" {
			params.Set("cursor", cursor)
		}
		var resp inventoryResponse
		if err := c.get(ctx, "/inventory", params, &resp); err != nil {
			return nil, "", err
		}
		return resp.Levels, resp.Cursor, nil
	})
}

// GetItemStock returns the current stock for the first variant of itemID.
// Returns 0 and no error when Loyverse has no inventory record for the item.
func (c *Client) GetItemStock(ctx context.Context, itemID string) (float64, error) {
	item, err := c.GetItem(ctx, itemID)
	if err != nil {
		return 0, fmt.Errorf("loyverse: get item stock: %w", err)
	}
	if len(item.Variants) == 0 {
		return 0, fmt.Errorf("loyverse: item %s has no variants", itemID)
	}

	params := url.Values{"variant_ids": {item.Variants[0].ID}}
	var resp inventoryResponse
	if err := c.get(ctx, "/inventory", params, &resp); err != nil {
		return 0, fmt.Errorf("loyverse: get inventory for item %s: %w", itemID, err)
	}
	if len(resp.Levels) == 0 {
		return 0, nil
	}
	return resp.Levels[0].InStock, nil
}

// SetStock sets the absolute stock level for a specific variant and store.
// Use this when you already know the variantID and storeID.
func (c *Client) SetStock(ctx context.Context, variantID, storeID string, stockAfter float64) error {
	body := map[string]any{
		"inventory_levels": []InventoryUpdate{
			{VariantID: variantID, StoreID: storeID, StockAfter: stockAfter},
		},
	}
	return c.post(ctx, "/inventory", body, nil)
}

// AdjustStock adds delta units to the current stock of itemID, resolving the variant
// and store automatically. Pass a negative delta to reduce stock.
//
// If no inventory record exists for the item (e.g. track_stock was just enabled),
// the adjustment is applied on top of a base stock of 0 using the item's first store.
func (c *Client) AdjustStock(ctx context.Context, itemID string, delta float64) error {
	item, err := c.GetItem(ctx, itemID)
	if err != nil {
		return fmt.Errorf("loyverse: adjust stock: %w", err)
	}
	if len(item.Variants) == 0 {
		return fmt.Errorf("loyverse: adjust stock: item %s has no variants", itemID)
	}
	variantID := item.Variants[0].ID

	params := url.Values{"variant_ids": {variantID}}
	var resp inventoryResponse
	if err := c.get(ctx, "/inventory", params, &resp); err != nil {
		return fmt.Errorf("loyverse: adjust stock: get inventory for item %s: %w", itemID, err)
	}

	var storeID string
	var current float64

	if len(resp.Levels) == 0 {
		// No inventory record yet; use the first store associated with the item.
		if len(item.Stores) == 0 {
			return fmt.Errorf("loyverse: adjust stock: item %s has no inventory record and no associated stores", itemID)
		}
		storeID = item.Stores[0].StoreID
		current = 0
	} else {
		storeID = resp.Levels[0].StoreID
		current = resp.Levels[0].InStock
	}

	return c.SetStock(ctx, variantID, storeID, current+delta)
}

// UpdateStockBatch sets the stock level for multiple variants in parallel.
// updates maps variantID → stockAfter absolute value.
//
// It fetches all current inventory levels once to resolve storeIDs, then dispatches
// up to c.workers concurrent POST /inventory requests.
// Returns (successful, failed) counts.
func (c *Client) UpdateStockBatch(ctx context.Context, updates map[string]float64) (ok, failed int, err error) {
	if len(updates) == 0 {
		return 0, 0, nil
	}

	levels, err := c.GetInventoryLevels(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("loyverse: batch update: fetch inventory levels: %w", err)
	}

	variantToStore := make(map[string]string, len(levels))
	for _, l := range levels {
		variantToStore[l.VariantID] = l.StoreID
	}

	type job struct {
		variantID string
		stock     float64
	}

	jobs := make(chan job, len(updates))
	for id, s := range updates {
		jobs <- job{variantID: id, stock: s}
	}
	close(jobs)

	var okCount, failCount int64
	var wg sync.WaitGroup

	for range c.workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				storeID, found := variantToStore[j.variantID]
				if !found {
					c.logger.ErrorContext(ctx, "loyverse: variant not in inventory levels, skipping",
						"variant_id", j.variantID,
					)
					atomic.AddInt64(&failCount, 1)
					continue
				}
				if setErr := c.SetStock(ctx, j.variantID, storeID, j.stock); setErr != nil {
					c.logger.ErrorContext(ctx, "loyverse: batch update failed",
						"variant_id", j.variantID,
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

// ResetCategoryStock sets the stock level of every variant of every item in categoryID to 0.
// It fetches all items once to identify variants in the category, then fetches all inventory
// levels once to resolve store associations. Runs up to c.workers concurrent requests.
// Returns (successful, failed) counts; per-variant errors are logged but do not abort the operation.
func (c *Client) ResetCategoryStock(ctx context.Context, categoryID string) (ok, failed int, err error) {
	items, err := c.GetItems(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("loyverse: reset category stock %s: fetch items: %w", categoryID, err)
	}

	variantSet := make(map[string]struct{})
	for _, item := range items {
		if item.CategoryID == categoryID {
			for _, v := range item.Variants {
				variantSet[v.ID] = struct{}{}
			}
		}
	}
	if len(variantSet) == 0 {
		return 0, 0, nil
	}

	levels, err := c.GetInventoryLevels(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("loyverse: reset category stock %s: fetch inventory: %w", categoryID, err)
	}

	var targets []InventoryLevel
	for _, l := range levels {
		if _, found := variantSet[l.VariantID]; found {
			targets = append(targets, l)
		}
	}
	if len(targets) == 0 {
		return 0, 0, nil
	}

	jobs := make(chan InventoryLevel, len(targets))
	for _, l := range targets {
		jobs <- l
	}
	close(jobs)

	var okCount, failCount int64
	var wg sync.WaitGroup

	for range c.workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for lvl := range jobs {
				if setErr := c.SetStock(ctx, lvl.VariantID, lvl.StoreID, 0); setErr != nil {
					c.logger.ErrorContext(ctx, "loyverse: reset category stock failed",
						"variant_id", lvl.VariantID,
						"category_id", categoryID,
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

// ResetAllStock sets the stock level of every tracked inventory record to 0.
// It fetches all inventory levels once and dispatches up to c.workers concurrent corrections.
// Returns (successful, failed) counts; per-variant errors are logged but do not abort the operation.
func (c *Client) ResetAllStock(ctx context.Context) (ok, failed int, err error) {
	levels, err := c.GetInventoryLevels(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("loyverse: reset all stock: %w", err)
	}
	if len(levels) == 0 {
		return 0, 0, nil
	}

	jobs := make(chan InventoryLevel, len(levels))
	for _, l := range levels {
		jobs <- l
	}
	close(jobs)

	var okCount, failCount int64
	var wg sync.WaitGroup

	for range c.workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for lvl := range jobs {
				if setErr := c.SetStock(ctx, lvl.VariantID, lvl.StoreID, 0); setErr != nil {
					c.logger.ErrorContext(ctx, "loyverse: reset all stock failed",
						"variant_id", lvl.VariantID,
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

// ResetNegativeStock sets all stock levels below zero back to zero.
// It fetches all inventory levels once and dispatches up to c.workers concurrent corrections.
// Returns (successful, failed) counts.
func (c *Client) ResetNegativeStock(ctx context.Context) (ok, failed int, err error) {
	levels, err := c.GetInventoryLevels(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("loyverse: reset negative stock: %w", err)
	}

	var negative []InventoryLevel
	for _, l := range levels {
		if l.InStock < 0 {
			negative = append(negative, l)
		}
	}
	if len(negative) == 0 {
		return 0, 0, nil
	}

	jobs := make(chan InventoryLevel, len(negative))
	for _, l := range negative {
		jobs <- l
	}
	close(jobs)

	var okCount, failCount int64
	var wg sync.WaitGroup

	for range c.workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for lvl := range jobs {
				if setErr := c.SetStock(ctx, lvl.VariantID, lvl.StoreID, 0); setErr != nil {
					c.logger.ErrorContext(ctx, "loyverse: reset negative stock failed",
						"variant_id", lvl.VariantID,
						"in_stock", lvl.InStock,
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
