package loyverse

import (
	"context"
	"fmt"
)

// ListPaymentTypes returns all payment types for the merchant account.
// The /payment_types endpoint is not cursor-paginated; all results are returned in a single response.
func (c *Client) ListPaymentTypes(ctx context.Context) ([]*PaymentType, error) {
	var resp paymentTypesResponse
	if err := c.get(ctx, "/payment_types", nil, &resp); err != nil {
		return nil, fmt.Errorf("loyverse: list payment types: %w", err)
	}
	ptrs := make([]*PaymentType, len(resp.PaymentTypes))
	for i := range resp.PaymentTypes {
		pt := resp.PaymentTypes[i]
		ptrs[i] = &pt
	}
	return ptrs, nil
}

// GetPaymentType returns a single payment type by its Loyverse ID.
func (c *Client) GetPaymentType(ctx context.Context, id string) (*PaymentType, error) {
	var pt PaymentType
	if err := c.get(ctx, "/payment_types/"+id, nil, &pt); err != nil {
		return nil, fmt.Errorf("loyverse: get payment type %s: %w", id, err)
	}
	return &pt, nil
}
