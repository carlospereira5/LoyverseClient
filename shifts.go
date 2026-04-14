package loyverse

import (
	"context"
	"net/url"
	"time"
)

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
