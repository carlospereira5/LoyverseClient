// Package webhook provides an [http.Handler] that receives Loyverse webhook events,
// verifies the HMAC-SHA256 signature, and delivers parsed receipts to a callback.
//
// The handler responds 200 to Loyverse immediately after parsing, then invokes
// the receipt callback asynchronously to avoid exhausting Loyverse's request timeout.
//
//	h := webhook.New(func(receipts []loyverse.Receipt) {
//	    for _, r := range receipts {
//	        // process each receipt
//	    }
//	}, webhook.WithSecret(os.Getenv("LOYVERSE_WEBHOOK_SECRET")))
//
//	http.Handle("/webhooks/loyverse", h)
package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/carlospereira5/loyverse"
)

// payload is the JSON structure sent by Loyverse for receipts.update events.
type payload struct {
	Receipts []loyverse.Receipt `json:"receipts"`
}

// Handler is an [http.Handler] for Loyverse inbound webhook events.
type Handler struct {
	secret    string
	onReceipt func(receipts []loyverse.Receipt)
	logger    *slog.Logger
}

// Option configures a Handler.
type Option func(*Handler)

// WithSecret sets the HMAC-SHA256 secret used to verify the X-Loyverse-Signature header.
// If not provided, signature verification is skipped (not recommended for production).
func WithSecret(secret string) Option {
	return func(h *Handler) { h.secret = secret }
}

// WithLogger sets the structured logger. Defaults to [slog.Default].
func WithLogger(l *slog.Logger) Option {
	return func(h *Handler) { h.logger = l }
}

// New creates a webhook Handler that delivers parsed receipts to onReceipt.
// onReceipt is invoked asynchronously in a new goroutine after the 200 response
// is sent, so it must not rely on the request lifecycle or context.
func New(onReceipt func(receipts []loyverse.Receipt), opts ...Option) *Handler {
	h := &Handler{
		onReceipt: onReceipt,
		logger:    slog.Default(),
	}
	for _, o := range opts {
		o(h)
	}
	return h
}

// ServeHTTP implements [http.Handler].
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("loyverse webhook: read request body", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if h.secret != "" {
		sig := r.Header.Get("X-Loyverse-Signature")
		if sig == "" || !h.verifySignature(body, sig) {
			h.logger.Warn("loyverse webhook: rejected request with invalid signature",
				"remote_addr", r.RemoteAddr,
			)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	} else {
		h.logger.Warn("loyverse webhook: signature verification disabled (no secret configured)")
	}

	var p payload
	if err := json.Unmarshal(body, &p); err != nil {
		h.logger.Error("loyverse webhook: parse payload", "err", err)
		// Respond 200 to prevent Loyverse from retrying a malformed payload indefinitely.
		w.WriteHeader(http.StatusOK)
		return
	}

	h.logger.Info("loyverse webhook: receipts received", "count", len(p.Receipts))

	// Respond 200 before invoking the callback so Loyverse's timeout is not exhausted
	// by application-side processing (DB writes, downstream API calls, etc.).
	w.WriteHeader(http.StatusOK)

	if len(p.Receipts) > 0 {
		receipts := p.Receipts // copy slice header before goroutine
		go h.onReceipt(receipts)
	}
}

// verifySignature returns true if HMAC-SHA256 of body under h.secret matches signature.
func (h *Handler) verifySignature(body []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(h.secret))
	_, _ = mac.Write(body) // hash.Hash.Write never returns a non-nil error
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}
