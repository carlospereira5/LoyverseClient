package loyverse

import "fmt"

// APIError is returned when the Loyverse API responds with a 4xx or 5xx status code.
// Use [errors.As] to extract StatusCode and Body for programmatic error handling.
//
//	var apiErr *loyverse.APIError
//	if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
//	    // handle not found
//	}
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("loyverse: API error %d: %s", e.StatusCode, e.Body)
}
