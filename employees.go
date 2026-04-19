package loyverse

import (
	"context"
	"fmt"
	"net/url"
)

// ListEmployees returns all employees for the merchant account, automatically
// following pagination cursors.
func (c *Client) ListEmployees(ctx context.Context) ([]Employee, error) {
	return paginate(func(cursor string) ([]Employee, string, error) {
		params := url.Values{"limit": {_pageLimit}}
		if cursor != "" {
			params.Set("cursor", cursor)
		}
		var resp employeesResponse
		if err := c.get(ctx, "/employees", params, &resp); err != nil {
			return nil, "", err
		}
		return resp.Employees, resp.Cursor, nil
	})
}

// GetEmployee returns a single employee by their Loyverse ID.
func (c *Client) GetEmployee(ctx context.Context, id string) (*Employee, error) {
	var e Employee
	if err := c.get(ctx, "/employees/"+id, nil, &e); err != nil {
		return nil, fmt.Errorf("loyverse: get employee %s: %w", id, err)
	}
	return &e, nil
}
