# loyverse

A Go client library for the [Loyverse API](https://developer.loyverse.com/).

## Installation

```bash
go get github.com/carlospereira5/loyverse
```

## Quick start

```go
import "github.com/carlospereira5/loyverse"

client, err := loyverse.New(os.Getenv("LOYVERSE_TOKEN"))
if err != nil {
    log.Fatal(err)
}

items, err := client.GetItems(ctx)
```

## Covered endpoints

| Resource     | Operations                                      |
|--------------|-------------------------------------------------|
| Items        | List, Get, Create, SetItemCost, ResetAllCosts   |
| Inventory    | List levels, GetItemStock, SetStock, AdjustStock, UpdateStockBatch, ResetNegativeStock |
| Receipts     | List (date range)                               |
| Shifts       | List (date range)                               |
| Categories   | List                                            |
| Webhook      | Inbound handler with HMAC-SHA256 verification   |

## Webhook

```go
import "github.com/carlospereira5/loyverse/webhook"

h := webhook.New(func(receipts []loyverse.Receipt) {
    // process receipts
}, webhook.WithSecret(os.Getenv("LOYVERSE_WEBHOOK_SECRET")))

http.Handle("/webhooks/loyverse", h)
```

## License

[MIT](LICENSE)
