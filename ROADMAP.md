# Roadmap

This document tracks planned work, in order of priority. Each version targets one coherent set of changes.
Completed items are marked ✅. Planned items are unmarked.

---

## v1.0.0 — Core client ✅ (released 2026-04-14)

Foundation: HTTP client, pagination, batch operations, and inbound webhook handler.

**Endpoints**
- ✅ `GET /items` — paginated catalog
- ✅ `GET /items/:id` — single item
- ✅ `POST /items` — create or update item
- ✅ `GET /inventory` — paginated stock levels
- ✅ `POST /inventory` — set absolute stock level
- ✅ `GET /receipts` — paginated list with date range
- ✅ `GET /shifts` — paginated list with date range
- ✅ `GET /categories` — paginated list

**Batch operations**
- ✅ `SetItemCost` — GET + modify + POST pattern; preserves all other fields
- ✅ `ResetAllCosts` — parallel worker pool over full catalog
- ✅ `AdjustStock` — delta-based adjustment with automatic variant/store resolution
- ✅ `UpdateStockBatch` — parallel stock updates from variantID → stockAfter map
- ✅ `ResetNegativeStock` — zero out all negative levels in parallel

**Webhook**
- ✅ `webhook.Handler` — HMAC-SHA256 signature verification
- ✅ Async callback invocation (responds 200 before processing)
- ✅ Configurable via functional options

**Infrastructure**
- ✅ `HTTPClient` interface for test injection
- ✅ Connection pooling via `http.Transport`
- ✅ `*slog.Logger` injection
- ✅ Configurable batch worker concurrency
- ✅ Custom base URL override (for tests)
- ✅ Full test suite with `httptest.Server` (zero real API calls)

---

## v1.1.0 — Stores and Variants ✅ (released 2026-04-18)

`store_id` is required by all inventory write operations but currently resolved as a side effect
of fetching item data. Exposing stores and standalone variants as first-class resources
eliminates unnecessary extra API calls.

**Endpoints**
- ✅ `GET /stores` — list all stores
- ✅ `GET /stores/:id` — single store
- ✅ `GET /variants` — standalone variant list (filter by barcode, SKU, or item ID)
- ✅ `GET /variants/:id` — single variant

**New types**: `Store`, `VariantStore`

**Also completed in this release**
- ✅ `GET /shifts/:id` — single shift
- ✅ `Shift` struct extended with full financial fields (`StartingCash`, `CashPayments`, `CashRefunds`, `PaidIn`, `ExpectedCash`, `ActualCash`, `GrossSales`, `Refunds`, `Discounts`, `NetSales`, `Tip`, `Surcharge`)
- ✅ New types: `ShiftPayment`, `ShiftTax`, `CashMovement`

---

## v1.2.0 — Read-only Resources ✅ (released 2026-04-19)

Low-effort additions that complete the read surface of the API.

**Endpoints**
- ✅ `GET /merchant` — merchant account information
- ✅ `GET /employees` — paginated employee list
- ✅ `GET /employees/:id` — single employee
- ✅ `GET /payment_types` — list payment types
- ✅ `GET /payment_types/:id` — single payment type

---

## v1.3.0 — Categories and Customers Write Operations ✅ (released 2026-04-19)

Extend the two most commonly mutated resources beyond read-only.

**Endpoints**
- ✅ `POST /categories` — create or update a category
- ✅ `DELETE /categories/:id` — delete a category
- ✅ `GET /customers` — paginated customer list
- ✅ `GET /customers/:id` — single customer
- ✅ `POST /customers` — create or update a customer
- ✅ `DELETE /customers/:id` — delete a customer

---

## v1.4.0 — Receipts Write Operations ✅ (released 2026-04-19)

Enables programmatic sales recording and refund issuance.

**Endpoints**
- ✅ `GET /receipts/:number` — single receipt by number
- ✅ `POST /receipts` — create a receipt
- ✅ `POST /receipts/:number/refund` — issue a refund against an existing receipt

**Also completed in this release**
- ✅ `Receipt` expanded with full API response fields (UpdatedAt, StoreID, EmployeeID, CustomerID, Tip, Surcharge, TotalTax, TotalDiscount, Payments, etc.)
- ✅ `LineItem` expanded with ID (for refund references), VariantName, SKU, calculated totals
- ✅ New types: `ReceiptPayment`, `PaymentDetails`, `CreateReceiptRequest`, `CreateReceiptLineItem`, `RefundRequest`, `RefundLineItem`

---

## v1.4.1 — Category Batch Operations (released 2026-04-22)

Targeted bulk mutations scoped to a single category.

**Batch operations**
- ✅ `ResetCategoryPrices` — zero the default price and per-store prices of all items in a given category
- ✅ `ResetCategoryStock` — zero the stock of all items in a given category

---

## v1.5.0 — Remaining Resources

Completes full API surface coverage for suppliers, taxes, discounts, modifiers, and POS devices.

**Endpoints**
- `GET /suppliers`, `GET /suppliers/:id`, `POST /suppliers`, `DELETE /suppliers/:id`
- `GET /taxes`, `GET /taxes/:id`, `POST /taxes`, `DELETE /taxes/:id`
- `GET /discounts`, `GET /discounts/:id`, `POST /discounts`, `DELETE /discounts/:id`
- `GET /modifiers`, `GET /modifiers/:id`, `POST /modifiers`, `DELETE /modifiers/:id`
- `GET /pos_devices`, `GET /pos_devices/:id`, `POST /pos_devices`, `DELETE /pos_devices/:id`
- `POST /items/:id/image`, `DELETE /items/:id/image` — item image management

---

## v2.0.0 — Robustness and Observability

Infrastructure improvements that make the client safe for high-traffic production use.

- Automatic retry with exponential backoff on `429 Too Many Requests` and `5xx` responses
- Configurable rate limiter to stay within Loyverse's per-minute API limits
- Request/response tracing hooks (optional `OnRequest`, `OnResponse` callbacks)
- GitHub Actions CI: `go test ./...`, `go vet ./...`, `go build ./...` on push
- Code coverage reporting
- Changelog and semantic versioning (`v2.0.0` tag)
