# Contributing to loyverse

This file defines the working conventions for this repository. All contributors
and automated agents must follow these rules without exception.

---

## Project scope

`github.com/carlospereira5/loyverse` is a Go client library for the Loyverse REST API.
Its only job is to translate Go types and method calls into authenticated HTTP requests
and parse the responses. No business logic, no application state, no framework coupling.

**Hard constraints:**
- Zero runtime dependencies. The `go.mod` must only contain `require` entries used
  in `*_test.go` files (currently `github.com/google/go-cmp`). If a proposed change
  requires a runtime dependency, the design must be reconsidered.
- Minimum Go version: `1.22`. Do not use language features introduced after `1.22`
  without updating this file and `go.mod` first.

---

## Roadmap

Planned work is tracked in [`ROADMAP.md`](ROADMAP.md).
Before starting any new feature, check that it is listed there and pick the next
unstarted item in version order. Do not skip versions or add out-of-scope features.

---

## Development workflow

### Branching

| Branch pattern       | Purpose                                  |
|----------------------|------------------------------------------|
| `master`             | Always stable. Never push directly.      |
| `feat/<scope>`       | New endpoint or feature                  |
| `fix/<scope>`        | Bug fix                                  |
| `test/<scope>`       | Test additions or improvements           |
| `chore/<scope>`      | Dependency updates, tooling, config      |
| `docs/<scope>`       | Documentation only                       |
| `refactor/<scope>`   | Code restructuring without behavior change |

### Commits

Use [Conventional Commits](https://www.conventionalcommits.org/) with a lowercase scope
matching the affected file or resource:

```
feat(items): add DELETE /items/:id endpoint
fix(inventory): handle missing store_id when track_stock is disabled
test(webhook): add test for empty receipts payload
chore(deps): update go-cmp to v0.7.1
docs(readme): add installation example
```

**Rules:**
- One logical unit per commit. A commit that adds an endpoint must not also refactor
  an existing one or fix an unrelated bug.
- Commit message body is optional but recommended when the change is non-obvious.
- Never reference issue numbers, AI tools, or external systems in commit messages.
- Do not amend published commits. Create a new commit to correct a mistake.

### Incremental development

Each version in the roadmap is implemented as a focused pull request:
one PR per version, one branch per PR. A PR may contain multiple commits
(one per endpoint or logical step) but must not span multiple roadmap versions.

Do not refactor existing code while adding a new feature. If a refactor is needed,
open a separate `refactor/` branch and merge it first.

---

## Code standards

### Package structure

The repository is a single Go module. Source files are organized by resource:

```
client.go          HTTP client, options, core helpers
types.go           All exported domain types
errors.go          APIError type
items.go           /items endpoints
inventory.go       /inventory endpoints
receipts.go        /receipts endpoints
shifts.go          /shifts endpoints
categories.go      /categories endpoints
webhook/
  webhook.go       Inbound webhook handler
```

New resources follow this pattern: one file per resource, named after the Loyverse
endpoint (`stores.go`, `variants.go`, etc.).

### Adding a new endpoint

1. Add the required types to `types.go` (or a new file if the type set is large).
2. Implement the endpoint method(s) on `*Client` in the resource file.
3. Write tests in the corresponding `*_test.go` file using `httptest.Server`.
4. Run `go vet ./...` and `go test ./...` before committing.

Follow existing patterns:
- Paginated list endpoints use the `paginate[T]` generic helper in `client.go`.
- Date-range queries use `formatDate` for consistent UTC formatting.
- All methods accept `context.Context` as the first parameter.
- Errors are wrapped with `fmt.Errorf("loyverse: <operation>: %w", err)`.
- API errors are returned as `*APIError` via the `do()` method.

### Error handling

- Never return a concrete error type from exported functions. Always return `error`.
- Wrap errors with context the caller does not have: resource name, ID, operation.
- Use `%w` to preserve the error chain for `errors.As` / `errors.Is`.
- Log at the call site (`do()`) only. Do not log and return simultaneously.

### Testing

- All tests use `package loyverse_test` (black-box). Access unexported symbols
  via `export_test.go` only when strictly necessary.
- Every new endpoint requires at minimum:
  - A happy-path test with a single-page response.
  - A pagination test (if the endpoint is paginated).
  - An `APIError` propagation test (server returns 4xx).
- Tests must not call the real Loyverse API. Use `httptest.NewServer` and
  `loyverse.WithBaseURL` to point the client at the test server.
- Shared setup lives in `testhelpers_test.go`. Do not duplicate fixture builders
  or server setup across test files.
- Use `cmp.Diff` from `github.com/google/go-cmp/cmp` for struct and slice comparisons.
- Use `t.Helper()` in every helper function that calls `t.Fatal` or `t.Error`.
- Use `t.Cleanup()` for teardown instead of `defer` inside helpers.

### Webhook package

The `webhook` package is intentionally minimal. It handles HMAC verification and
JSON parsing only. It must never contain application logic (session management,
database writes, TUI messaging, etc.). Those concerns belong in the consuming
application, not in this library.

---

## What this library is not

- It is not an ORM or query builder for Loyverse data.
- It is not a sync engine or cache layer.
- It does not manage authentication flows (token generation is out of scope).
- It does not include retry logic (planned for v2.0.0; do not add it earlier).
