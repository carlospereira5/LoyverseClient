package loyverse

import "time"

// Store represents a physical or virtual store location in a Loyverse merchant account.
type Store struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Address     string     `json:"address"`
	PhoneNumber string     `json:"phone_number"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
}

// Item represents a product in the Loyverse catalog.
type Item struct {
	ID         string      `json:"id"`
	Name       string      `json:"item_name"`
	CategoryID string      `json:"category_id"`
	TrackStock bool        `json:"track_stock"`
	Cost       float64     `json:"cost"`
	Variants   []Variant   `json:"variants"`
	Stores     []ItemStore `json:"stores"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	DeletedAt  *time.Time  `json:"deleted_at"`
}

// Variant represents a product variant within an Item.
// When retrieved via the standalone /variants endpoint additional fields are populated.
type Variant struct {
	ID           string  `json:"variant_id"`
	ItemID       string  `json:"item_id"`
	Name         string  `json:"name"`
	SKU          string  `json:"sku"`
	Barcode      string  `json:"barcode"`
	Cost         float64 `json:"cost"`
	PurchaseCost float64 `json:"purchase_cost"`
	// DefaultPrice is the sale price. May be 0 when PricingType is "VARIABLE".
	DefaultPrice float64 `json:"default_price"`
	// PricingType is "FIXED" or "VARIABLE".
	PricingType string `json:"default_pricing_type"`

	// Fields populated only by the standalone /variants endpoint.
	ReferenceVariantID string         `json:"reference_variant_id,omitempty"`
	Option1Value       string         `json:"option1_value,omitempty"`
	Option2Value       string         `json:"option2_value,omitempty"`
	Option3Value       string         `json:"option3_value,omitempty"`
	Stores             []VariantStore `json:"stores,omitempty"`
	CreatedAt          time.Time      `json:"created_at,omitempty"`
	UpdatedAt          time.Time      `json:"updated_at,omitempty"`
	DeletedAt          *time.Time     `json:"deleted_at,omitempty"`
}

// VariantStore holds store-specific pricing and availability for a Variant.
type VariantStore struct {
	StoreID          string   `json:"store_id"`
	PricingType      string   `json:"pricing_type"`
	Price            *float64 `json:"price"`
	AvailableForSale bool     `json:"available_for_sale"`
	OptimalStock     *float64 `json:"optimal_stock"`
	LowStock         *float64 `json:"low_stock"`
}

// ItemStore holds store-specific data associated with an Item.
type ItemStore struct {
	StoreID string `json:"store_id"`
}

// Category represents a product category.
type Category struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// CategoryRequest is the body for POST /categories (create or update).
// Set ID to update an existing category; omit it to create a new one.
type CategoryRequest struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

// Receipt represents a sale or refund transaction recorded by a POS device.
type Receipt struct {
	ReceiptNumber string `json:"receipt_number"`
	// ReceiptType is "SALE" or "REFUND".
	ReceiptType string `json:"receipt_type"`
	// Status is "DONE", "CANCELLED", or "OPEN".
	Status     string  `json:"status"`
	TotalMoney float64 `json:"total_money"`
	// ReceiptDate is the actual transaction time on the POS device.
	// For offline sales this differs from CreatedAt (server upload time).
	// Loyverse uses ReceiptDate in its Back Office reports.
	ReceiptDate time.Time `json:"receipt_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// CancelledAt is nil for non-cancelled receipts.
	CancelledAt *time.Time `json:"cancelled_at"`

	// Optional context fields.
	Note        string `json:"note,omitempty"`
	// RefundFor holds the receipt number of the original sale for REFUND receipts.
	RefundFor   string `json:"refund_for,omitempty"`
	Order       string `json:"order,omitempty"`
	Source      string `json:"source,omitempty"`
	EmployeeID  string `json:"employee_id,omitempty"`
	CustomerID  string `json:"customer_id,omitempty"`
	StoreID     string `json:"store_id,omitempty"`
	POSDeviceID string `json:"pos_device_id,omitempty"`

	// Financial totals.
	TotalTax      float64 `json:"total_tax,omitempty"`
	TotalDiscount float64 `json:"total_discount,omitempty"`
	Tip           float64 `json:"tip,omitempty"`
	Surcharge     float64 `json:"surcharge,omitempty"`

	// Collections.
	LineItems []LineItem       `json:"line_items"`
	Payments  []ReceiptPayment `json:"payments,omitempty"`
}

// LineItem represents a single product line within a Receipt.
type LineItem struct {
	// ID is the line item's identifier within the receipt; used to reference it in refunds.
	ID          string  `json:"id,omitempty"`
	ItemID      string  `json:"item_id"`
	ItemName    string  `json:"item_name"`
	VariantID   string  `json:"variant_id"`
	VariantName string  `json:"variant_name,omitempty"`
	SKU         string  `json:"sku,omitempty"`
	Quantity    float64 `json:"quantity"`
	Price       float64 `json:"price"`

	// Calculated totals.
	GrossTotalMoney float64 `json:"gross_total_money,omitempty"`
	TotalMoney      float64 `json:"total_money,omitempty"`
	Cost            float64 `json:"cost,omitempty"`
	CostTotal       float64 `json:"cost_total,omitempty"`
	TotalDiscount   float64 `json:"total_discount,omitempty"`
	LineNote        string  `json:"line_note,omitempty"`
}

// ReceiptPayment represents a payment applied to a receipt.
type ReceiptPayment struct {
	PaymentTypeID  string          `json:"payment_type_id,omitempty"`
	MoneyAmount    float64         `json:"money_amount"`
	Name           string          `json:"name,omitempty"`
	Type           string          `json:"type,omitempty"`
	PaidAt         *time.Time      `json:"paid_at,omitempty"`
	PaymentDetails *PaymentDetails `json:"payment_details,omitempty"`
}

// PaymentDetails holds card-specific information for a ReceiptPayment.
type PaymentDetails struct {
	AuthorizationCode string `json:"authorization_code,omitempty"`
	ReferenceID       string `json:"reference_id,omitempty"`
	EntryMethod       string `json:"entry_method,omitempty"`
	CardCompany       string `json:"card_company,omitempty"`
	CardNumber        string `json:"card_number,omitempty"`
}

// ShiftTax holds the money collected for a single tax during a shift.
type ShiftTax struct {
	TaxID       string  `json:"tax_id"`
	MoneyAmount float64 `json:"money_amount"`
}

// ShiftPayment holds the total money collected via one payment type during a shift.
type ShiftPayment struct {
	PaymentTypeID string  `json:"payment_type_id"`
	MoneyAmount   float64 `json:"money_amount"`
}

// CashMovement represents a single paid-in or paid-out cash movement within a shift.
type CashMovement struct {
	// Type is "PAID_IN" or "PAID_OUT".
	Type        string    `json:"type"`
	MoneyAmount float64   `json:"money_amount"`
	Comment     string    `json:"comment,omitempty"`
	EmployeeID  string    `json:"employee_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Shift represents a cash register shift.
type Shift struct {
	ID           string `json:"id"`
	StoreID      string `json:"store_id,omitempty"`
	POSDeviceID  string `json:"pos_device_id,omitempty"`
	OpenedAt     time.Time `json:"opened_at"`
	// ClosedAt is nil for currently open shifts.
	ClosedAt          *time.Time `json:"closed_at"`
	OpenedByEmployee  string     `json:"opened_by_employee,omitempty"`
	ClosedByEmployee  string     `json:"closed_by_employee,omitempty"`

	// Cash summary fields.
	StartingCash float64 `json:"starting_cash"`
	CashPayments float64 `json:"cash_payments"`
	CashRefunds  float64 `json:"cash_refunds"`
	PaidIn       float64 `json:"paid_in"`
	PaidOut      float64 `json:"paid_out"`
	ExpectedCash float64 `json:"expected_cash"`
	ActualCash   float64 `json:"actual_cash"`

	// Sales summary fields.
	GrossSales float64 `json:"gross_sales"`
	Refunds    float64 `json:"refunds"`
	Discounts  float64 `json:"discounts"`
	NetSales   float64 `json:"net_sales"`
	Tip        float64 `json:"tip"`
	Surcharge  float64 `json:"surcharge"`

	// Collections.
	Taxes         []ShiftTax     `json:"taxes,omitempty"`
	Payments      []ShiftPayment `json:"payments,omitempty"`
	CashMovements []CashMovement `json:"cash_movements,omitempty"`
}

// InventoryLevel represents the current stock level of a variant in a specific store.
type InventoryLevel struct {
	VariantID string  `json:"variant_id"`
	StoreID   string  `json:"store_id"`
	InStock   float64 `json:"in_stock"`
}

// InventoryUpdate is a single stock adjustment entry used in POST /inventory requests.
type InventoryUpdate struct {
	VariantID  string  `json:"variant_id"`
	StoreID    string  `json:"store_id"`
	StockAfter float64 `json:"stock_after"`
}

// CreateItemRequest is the payload for creating a new item via POST /items.
type CreateItemRequest struct {
	Name       string                 `json:"item_name"`
	CategoryID string                 `json:"category_id,omitempty"`
	TrackStock bool                   `json:"track_stock"`
	Cost       float64                `json:"cost,omitempty"`
	Variants   []CreateVariantRequest `json:"variants"`
}

// CreateVariantRequest is the variant definition within a [CreateItemRequest].
type CreateVariantRequest struct {
	// PricingType must be "FIXED" for DefaultPrice to take effect.
	// Use "VARIABLE" for open pricing at the point of sale.
	PricingType  string  `json:"default_pricing_type"`
	DefaultPrice float64 `json:"default_price,omitempty"`
	Barcode      string  `json:"barcode,omitempty"`
	SKU          string  `json:"sku,omitempty"`
}

// Merchant represents the Loyverse merchant account.
type Merchant struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	CurrencyCode string    `json:"currency_code"`
	LanguageCode string    `json:"language_code"`
	CountryCode  string    `json:"country_code"`
	Address      string    `json:"address"`
	PhoneNumber  string    `json:"phone_number"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Employee represents a staff member in a Loyverse merchant account.
type Employee struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email,omitempty"`
	PhoneNumber string     `json:"phone_number,omitempty"`
	// Stores holds the IDs of stores the employee has access to.
	Stores    []string   `json:"stores"`
	IsOwner   bool       `json:"is_owner"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// PaymentType represents a payment method configured in a Loyverse merchant account.
type PaymentType struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	// Type is "CASH", "CARD", or a custom type string.
	Type      string     `json:"type"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// Customer represents a customer in a Loyverse merchant account.
type Customer struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	PhoneNumber  string     `json:"phone_number"`
	Address      string     `json:"address"`
	City         string     `json:"city"`
	Region       string     `json:"region"`
	PostalCode   string     `json:"postal_code"`
	CountryCode  string     `json:"country_code"`
	CustomerCode string     `json:"customer_code"`
	Note         string     `json:"note"`
	FirstVisit   *time.Time `json:"first_visit"`
	LastVisit    *time.Time `json:"last_visit"`
	TotalVisits  int        `json:"total_visits"`
	TotalSpent   float64    `json:"total_spent"`
	TotalPoints  float64    `json:"total_points"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CreateReceiptRequest is the payload for POST /receipts.
type CreateReceiptRequest struct {
	StoreID     string                  `json:"store_id"`
	EmployeeID  string                  `json:"employee_id,omitempty"`
	CustomerID  string                  `json:"customer_id,omitempty"`
	Source      string                  `json:"source,omitempty"`
	Order       string                  `json:"order,omitempty"`
	ReceiptDate *time.Time              `json:"receipt_date,omitempty"`
	Note        string                  `json:"note,omitempty"`
	LineItems   []CreateReceiptLineItem `json:"line_items"`
	Payments    []ReceiptPayment        `json:"payments,omitempty"`
}

// CreateReceiptLineItem is a line item within a [CreateReceiptRequest].
type CreateReceiptLineItem struct {
	VariantID     string                `json:"variant_id"`
	Quantity      float64               `json:"quantity"`
	Price         float64               `json:"price,omitempty"`
	Cost          float64               `json:"cost,omitempty"`
	LineNote      string                `json:"line_note,omitempty"`
	LineTaxes     []LineItemTaxRef      `json:"line_taxes,omitempty"`
	LineDiscounts []LineItemDiscountRef  `json:"line_discounts,omitempty"`
	LineModifiers []LineItemModifierRef  `json:"line_modifiers,omitempty"`
}

// LineItemTaxRef references a tax to apply to a CreateReceiptLineItem.
type LineItemTaxRef struct {
	ID string `json:"id"`
}

// LineItemDiscountRef references a discount to apply to a CreateReceiptLineItem.
type LineItemDiscountRef struct {
	ID string `json:"id"`
}

// LineItemModifierRef references a modifier option to apply to a CreateReceiptLineItem.
type LineItemModifierRef struct {
	ModifierOptionID string `json:"modifier_option_id"`
}

// CustomerRequest is the body for POST /customers (create or update).
// Set ID to update an existing customer; omit it to create a new one.
type CustomerRequest struct {
	ID           string `json:"id,omitempty"`
	Name         string `json:"name"`
	Email        string `json:"email,omitempty"`
	PhoneNumber  string `json:"phone_number,omitempty"`
	Address      string `json:"address,omitempty"`
	City         string `json:"city,omitempty"`
	Region       string `json:"region,omitempty"`
	PostalCode   string `json:"postal_code,omitempty"`
	CountryCode  string `json:"country_code,omitempty"`
	CustomerCode string `json:"customer_code,omitempty"`
	Note         string `json:"note,omitempty"`
}

// --- internal response envelopes (not exported) ---

type itemsResponse struct {
	Items  []Item `json:"items"`
	Cursor string `json:"cursor"`
}

type receiptsResponse struct {
	Receipts []Receipt `json:"receipts"`
	Cursor   string    `json:"cursor"`
}

type shiftsResponse struct {
	Shifts []Shift `json:"shifts"`
	Cursor string  `json:"cursor"`
}

type categoriesResponse struct {
	Categories []Category `json:"categories"`
	Cursor     string     `json:"cursor"`
}

type inventoryResponse struct {
	Levels []InventoryLevel `json:"inventory_levels"`
	Cursor string           `json:"cursor"`
}

type storesResponse struct {
	Stores []Store `json:"stores"`
}

type variantsResponse struct {
	Variants []Variant `json:"variants"`
	Cursor   string    `json:"cursor"`
}

type employeesResponse struct {
	Employees []Employee `json:"employees"`
	Cursor    string     `json:"cursor"`
}

type paymentTypesResponse struct {
	PaymentTypes []PaymentType `json:"payment_types"`
}

type customersResponse struct {
	Customers []Customer `json:"customers"`
	Cursor    string     `json:"cursor"`
}
