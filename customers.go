package loyverse

import (
	"context"
	"fmt"
	"net/url"
)

// ListCustomers returns all customers, automatically following pagination cursors.
func (c *Client) ListCustomers(ctx context.Context) ([]*Customer, error) {
	return paginate(func(cursor string) ([]*Customer, string, error) {
		params := url.Values{"limit": {_pageLimit}}
		if cursor != "" {
			params.Set("cursor", cursor)
		}
		var resp customersResponse
		if err := c.get(ctx, "/customers", params, &resp); err != nil {
			return nil, "", err
		}
		return resp.Customers, resp.Cursor, nil
	})
}

// GetCustomer returns a single customer by its Loyverse ID.
func (c *Client) GetCustomer(ctx context.Context, id string) (*Customer, error) {
	var cust Customer
	if err := c.get(ctx, "/customers/"+id, nil, &cust); err != nil {
		return nil, fmt.Errorf("loyverse: get customer %s: %w", id, err)
	}
	return &cust, nil
}

// CreateOrUpdateCustomer creates a new customer or updates an existing one via POST /customers.
// Set req.ID to update; omit it to create.
func (c *Client) CreateOrUpdateCustomer(ctx context.Context, req CustomerRequest) (*Customer, error) {
	var cust Customer
	if err := c.post(ctx, "/customers", req, &cust); err != nil {
		return nil, fmt.Errorf("loyverse: create or update customer: %w", err)
	}
	return &cust, nil
}

// DeleteCustomer permanently deletes the customer with the given ID via DELETE /customers/:id.
func (c *Client) DeleteCustomer(ctx context.Context, id string) error {
	if err := c.delete(ctx, "/customers/"+id); err != nil {
		return fmt.Errorf("loyverse: delete customer %s: %w", id, err)
	}
	return nil
}
