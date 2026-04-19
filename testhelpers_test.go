package loyverse_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/carlospereira5/loyverse"
)

// newTestClient starts an httptest.Server backed by mux and returns a Client pointed at it.
// The server is shut down automatically when the test ends via t.Cleanup.
func newTestClient(t *testing.T, mux *http.ServeMux) *loyverse.Client {
	t.Helper()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c, err := loyverse.New("test-token", loyverse.WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("loyverse.New() error = %v", err)
	}
	return c
}

// mustWriteJSON encodes v as JSON and writes it to w with the correct Content-Type header.
func mustWriteJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Errorf("mustWriteJSON: %v", err)
	}
}

// --- fixtures ---

func itemFixture() loyverse.Item {
	return loyverse.Item{
		ID:   "item-1",
		Name: "Test Item",
		Variants: []loyverse.Variant{
			{ID: "var-1", ItemID: "item-1", Barcode: "123456"},
		},
		Stores: []loyverse.ItemStore{
			{StoreID: "store-1"},
		},
	}
}

func inventoryLevelFixture() loyverse.InventoryLevel {
	return loyverse.InventoryLevel{
		VariantID: "var-1",
		StoreID:   "store-1",
		InStock:   10,
	}
}

func receiptFixture() loyverse.Receipt {
	ts := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	return loyverse.Receipt{
		ReceiptNumber: "R-001",
		ReceiptType:   "SALE",
		Status:        "DONE",
		TotalMoney:    100.0,
		ReceiptDate:   ts,
		CreatedAt:     ts,
		LineItems: []loyverse.LineItem{
			{ItemID: "item-1", ItemName: "Test Item", Quantity: 2, Price: 50.0},
		},
	}
}

func categoryFixture() loyverse.Category {
	return loyverse.Category{
		ID:   "cat-1",
		Name: "Beverages",
	}
}

func paymentTypeFixture() loyverse.PaymentType {
	ts := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	return loyverse.PaymentType{
		ID:        "pt-1",
		Name:      "Cash",
		Type:      "CASH",
		CreatedAt: ts,
		UpdatedAt: ts,
	}
}

func employeeFixture() loyverse.Employee {
	ts := time.Date(2025, 2, 1, 9, 0, 0, 0, time.UTC)
	return loyverse.Employee{
		ID:          "emp-1",
		Name:        "Jane Smith",
		Email:       "jane@acme.com",
		PhoneNumber: "+1-555-0200",
		Stores:      []string{"store-1"},
		IsOwner:     false,
		CreatedAt:   ts,
		UpdatedAt:   ts,
	}
}

func merchantFixture() loyverse.Merchant {
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	return loyverse.Merchant{
		ID:           "merchant-1",
		Name:         "Acme Corp",
		Email:        "owner@acme.com",
		CurrencyCode: "USD",
		LanguageCode: "en",
		CountryCode:  "US",
		Address:      "1 Main St",
		PhoneNumber:  "+1-555-0100",
		CreatedAt:    ts,
		UpdatedAt:    ts,
	}
}

func customerFixture() loyverse.Customer {
	ts := time.Date(2025, 3, 1, 10, 0, 0, 0, time.UTC)
	fv := time.Date(2025, 3, 5, 0, 0, 0, 0, time.UTC)
	lv := time.Date(2025, 4, 10, 0, 0, 0, 0, time.UTC)
	return loyverse.Customer{
		ID:           "cust-1",
		Name:         "Alice Wonderland",
		Email:        "alice@example.com",
		PhoneNumber:  "+1-555-0300",
		Address:      "99 Rabbit Hole Lane",
		City:         "Springfield",
		Region:       "IL",
		PostalCode:   "62701",
		CountryCode:  "US",
		CustomerCode: "ALICE01",
		Note:         "VIP customer",
		FirstVisit:   &fv,
		LastVisit:    &lv,
		TotalVisits:  12,
		TotalSpent:   540.0,
		TotalPoints:  540,
		CreatedAt:    ts,
		UpdatedAt:    ts,
	}
}

func shiftFixture() loyverse.Shift {
	opened := time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC)
	closed := time.Date(2025, 1, 15, 20, 0, 0, 0, time.UTC)
	movementAt := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	return loyverse.Shift{
		ID:               "shift-1",
		StoreID:          "store-1",
		POSDeviceID:      "pos-1",
		OpenedAt:         opened,
		ClosedAt:         &closed,
		OpenedByEmployee: "emp-1",
		ClosedByEmployee: "emp-2",
		StartingCash:     200.0,
		CashPayments:     1500.0,
		CashRefunds:      50.0,
		PaidIn:           100.0,
		PaidOut:          500.0,
		ExpectedCash:     1250.0,
		ActualCash:       1240.0,
		GrossSales:       2000.0,
		Refunds:          50.0,
		Discounts:        30.0,
		NetSales:         1920.0,
		Tip:              20.0,
		Surcharge:        5.0,
		Taxes: []loyverse.ShiftTax{
			{TaxID: "tax-1", MoneyAmount: 192.0},
		},
		Payments: []loyverse.ShiftPayment{
			{PaymentTypeID: "pt-cash", MoneyAmount: 1500.0},
			{PaymentTypeID: "pt-card", MoneyAmount: 500.0},
		},
		CashMovements: []loyverse.CashMovement{
			{
				Type:        "PAID_OUT",
				MoneyAmount: 500.0,
				Comment:     "end of day float",
				EmployeeID:  "emp-1",
				CreatedAt:   movementAt,
			},
		},
	}
}
