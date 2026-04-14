package loyverse

import (
	"context"
	"net/url"
	"time"
)

// ListReceipts returns all receipts with created_at between since and until (inclusive),
// automatically following pagination cursors.
//
// Both bounds are interpreted as UTC created_at timestamps.
// For offline POS sales, use Receipt.ReceiptDate for the actual transaction time.
func (c *Client) ListReceipts(ctx context.Context, since, until time.Time) ([]Receipt, error) {
	return paginate(func(cursor string) ([]Receipt, string, error) {
		params := url.Values{
			"limit":          {_pageLimit},
			"created_at_min": {formatDate(since)},
			"created_at_max": {formatDate(until)},
		}
		if cursor != "" {
			params.Set("cursor", cursor)
		}
		var resp receiptsResponse
		if err := c.get(ctx, "/receipts", params, &resp); err != nil {
			return nil, "", err
		}
		return resp.Receipts, resp.Cursor, nil
	})
}
