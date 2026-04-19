package loyverse

import (
	"context"
	"fmt"
	"net/url"
)

// VariantsFilter holds optional filter parameters for [Client.ListVariants].
// All fields are optional; zero values are ignored.
type VariantsFilter struct {
	// VariantIDs is a comma-separated list of variant IDs to return.
	VariantIDs string
	// ItemIDs is a comma-separated list of item IDs; returns only variants attached to those items.
	ItemIDs string
	// SKU filters variants by exact SKU match.
	SKU string
}

// ListVariants returns all variants, automatically following pagination cursors.
// Pass a zero-value [VariantsFilter] to return all variants without filtering.
func (c *Client) ListVariants(ctx context.Context, f VariantsFilter) ([]Variant, error) {
	return paginate(func(cursor string) ([]Variant, string, error) {
		params := url.Values{"limit": {_pageLimit}}
		if cursor != "" {
			params.Set("cursor", cursor)
		}
		if f.VariantIDs != "" {
			params.Set("variants_ids", f.VariantIDs)
		}
		if f.ItemIDs != "" {
			params.Set("items_ids", f.ItemIDs)
		}
		if f.SKU != "" {
			params.Set("sku", f.SKU)
		}
		var resp variantsResponse
		if err := c.get(ctx, "/variants", params, &resp); err != nil {
			return nil, "", err
		}
		return resp.Variants, resp.Cursor, nil
	})
}

// GetVariant returns a single variant by its Loyverse ID.
func (c *Client) GetVariant(ctx context.Context, id string) (*Variant, error) {
	var v Variant
	if err := c.get(ctx, "/variants/"+id, nil, &v); err != nil {
		return nil, fmt.Errorf("loyverse: get variant %s: %w", id, err)
	}
	return &v, nil
}
