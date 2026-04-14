package webhook_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/carlospereira5/loyverse"
	"github.com/carlospereira5/loyverse/webhook"
)

const testSecret = "test-webhook-secret"

// signBody computes the HMAC-SHA256 signature that Loyverse sends in X-Loyverse-Signature.
func signBody(t *testing.T, secret, body string) string {
	t.Helper()
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// postWebhook sends a POST request to h with the given body and optional signature header.
func postWebhook(t *testing.T, h http.Handler, body, signature string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if signature != "" {
		req.Header.Set("X-Loyverse-Signature", signature)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// receiptPayload builds a minimal receipts.update JSON payload.
func receiptPayload(t *testing.T, receipts []loyverse.Receipt) string {
	t.Helper()
	b, err := json.Marshal(map[string]any{"receipts": receipts})
	if err != nil {
		t.Fatalf("receiptPayload: %v", err)
	}
	return string(b)
}

// --- tests ---

func TestHandler_validSignature_callbackReceivesReceipts(t *testing.T) {
	receipt := loyverse.Receipt{
		ReceiptNumber: "R-001",
		ReceiptType:   "SALE",
		Status:        "DONE",
	}
	body := receiptPayload(t, []loyverse.Receipt{receipt})
	sig := signBody(t, testSecret, body)

	done := make(chan []loyverse.Receipt, 1)
	h := webhook.New(func(receipts []loyverse.Receipt) {
		done <- receipts
	}, webhook.WithSecret(testSecret))

	rec := postWebhook(t, h, body, sig)

	if rec.Code != http.StatusOK {
		t.Errorf("ServeHTTP status = %d, want %d", rec.Code, http.StatusOK)
	}

	select {
	case got := <-done:
		if len(got) != 1 {
			t.Fatalf("callback received %d receipts, want 1", len(got))
		}
		if got[0].ReceiptNumber != receipt.ReceiptNumber {
			t.Errorf("callback receipt.ReceiptNumber = %q, want %q", got[0].ReceiptNumber, receipt.ReceiptNumber)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("callback was not invoked within 2 seconds")
	}
}

func TestHandler_invalidSignature_returns401(t *testing.T) {
	body := receiptPayload(t, []loyverse.Receipt{{ReceiptNumber: "R-001"}})
	callbackCalled := false

	h := webhook.New(func(receipts []loyverse.Receipt) {
		callbackCalled = true
	}, webhook.WithSecret(testSecret))

	rec := postWebhook(t, h, body, "invalid-signature")

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("ServeHTTP status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if callbackCalled {
		t.Error("callback was invoked on invalid signature, want no invocation")
	}
}

func TestHandler_missingSignature_returns401(t *testing.T) {
	body := receiptPayload(t, []loyverse.Receipt{{ReceiptNumber: "R-001"}})
	callbackCalled := false

	h := webhook.New(func(receipts []loyverse.Receipt) {
		callbackCalled = true
	}, webhook.WithSecret(testSecret))

	rec := postWebhook(t, h, body, "" /* no signature header */)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("ServeHTTP status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if callbackCalled {
		t.Error("callback was invoked on missing signature, want no invocation")
	}
}

func TestHandler_noSecret_skipsVerification(t *testing.T) {
	body := receiptPayload(t, []loyverse.Receipt{{ReceiptNumber: "R-001", ReceiptType: "SALE"}})

	done := make(chan struct{}, 1)
	h := webhook.New(func(receipts []loyverse.Receipt) {
		done <- struct{}{}
	}) // no WithSecret — verification disabled

	rec := postWebhook(t, h, body, "" /* no signature */)

	if rec.Code != http.StatusOK {
		t.Errorf("ServeHTTP status = %d, want %d", rec.Code, http.StatusOK)
	}
	select {
	case <-done:
		// callback invoked as expected
	case <-time.After(2 * time.Second):
		t.Fatal("callback was not invoked within 2 seconds")
	}
}

func TestHandler_emptyReceipts_callbackNotInvoked(t *testing.T) {
	body := receiptPayload(t, []loyverse.Receipt{})
	sig := signBody(t, testSecret, body)

	callbackCalled := false
	h := webhook.New(func(receipts []loyverse.Receipt) {
		callbackCalled = true
	}, webhook.WithSecret(testSecret))

	rec := postWebhook(t, h, body, sig)

	if rec.Code != http.StatusOK {
		t.Errorf("ServeHTTP status = %d, want %d", rec.Code, http.StatusOK)
	}
	// Give the goroutine time to run if it was (incorrectly) spawned.
	time.Sleep(50 * time.Millisecond)
	if callbackCalled {
		t.Error("callback was invoked on empty receipts, want no invocation")
	}
}

func TestHandler_malformedJSON_returns200(t *testing.T) {
	body := `{this is not valid json`
	sig := signBody(t, testSecret, body)

	callbackCalled := false
	h := webhook.New(func(receipts []loyverse.Receipt) {
		callbackCalled = true
	}, webhook.WithSecret(testSecret))

	rec := postWebhook(t, h, body, sig)

	// Must respond 200 to prevent Loyverse from retrying a permanently malformed payload.
	if rec.Code != http.StatusOK {
		t.Errorf("ServeHTTP status = %d, want %d (must not retry malformed payload)", rec.Code, http.StatusOK)
	}
	if callbackCalled {
		t.Error("callback was invoked on malformed JSON, want no invocation")
	}
}

func TestHandler_wrongMethod_returns405(t *testing.T) {
	h := webhook.New(func(receipts []loyverse.Receipt) {})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("ServeHTTP GET status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}
