package loyverse

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// GetShift returns a single shift by its Loyverse ID.
func (c *Client) GetShift(ctx context.Context, id string) (*Shift, error) {
	var shift Shift
	if err := c.get(ctx, "/shifts/"+id, nil, &shift); err != nil {
		return nil, fmt.Errorf("loyverse: get shift %s: %w", id, err)
	}
	return &shift, nil
}

// ListShifts returns all shifts opened between since and until (inclusive),
// automatically following pagination cursors.
func (c *Client) ListShifts(ctx context.Context, since, until time.Time) ([]Shift, error) {
	return paginate(func(cursor string) ([]Shift, string, error) {
		params := url.Values{
			"limit":          {_pageLimit},
			"created_at_min": {formatDate(since)},
			"created_at_max": {formatDate(until)},
		}
		if cursor != "" {
			params.Set("cursor", cursor)
		}
		var resp shiftsResponse
		if err := c.get(ctx, "/shifts", params, &resp); err != nil {
			return nil, "", err
		}
		return resp.Shifts, resp.Cursor, nil
	})
}
