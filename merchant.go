package loyverse

import (
	"context"
	"fmt"
)

// GetMerchant returns the merchant account information for the authenticated token.
// The /merchant endpoint returns a single object with no ID path parameter.
func (c *Client) GetMerchant(ctx context.Context) (*Merchant, error) {
	var m Merchant
	if err := c.get(ctx, "/merchant", nil, &m); err != nil {
		return nil, fmt.Errorf("loyverse: get merchant: %w", err)
	}
	return &m, nil
}
