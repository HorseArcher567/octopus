package mapstruct

import (
	"testing"
	"time"
)

// ===== Business scenario: e-commerce order =====

// Basic user info
type UserInfo struct {
	ID       int
	Username string
	Email    string
	Phone    string
}

// Address info
type Address struct {
	Street     string
	City       string
	State      string
	PostalCode string
	Country    string
}

// Product info
type Product struct {
	ID          int
	Name        string
	Price       float64
	Category    string
	Description string
}

// Order item
type OrderItem struct {
	Product  Product
	Quantity int
	Subtotal float64
}

// Payment info
type PaymentInfo struct {
	Method      string
	Amount      float64
	Currency    string
	Transaction string
	Status      string
}

// Shipping info
type ShippingInfo struct {
	Address     Address
	Method      string
	TrackingNum string
	Estimated   string
}

// Order status
type OrderStatus struct {
	Status    string
	UpdatedAt time.Time
	Notes     string
}

// Full order (5-level nested structure)
type Order struct {
	User        UserInfo
	Items       []OrderItem
	Payment     PaymentInfo
	Shipping    ShippingInfo
	OrderStatus OrderStatus
	OrderID     string
	Total       float64
	CreatedAt   time.Time
}

// ===== Test cases =====

func TestECommerceOrderDecoding(t *testing.T) {
	decoder := New()

	t.Run("Complete Order Decoding", func(t *testing.T) {
		input := map[string]any{
			"User": map[string]any{
				"ID":       1001,
				"Username": "john_doe",
				"Email":    "john@example.com",
				"Phone":    "+1-555-0123",
			},
			"Items": []map[string]any{
				{
					"Product": map[string]any{
						"ID":          101,
						"Name":        "MacBook Pro",
						"Price":       1999.99,
						"Category":    "Electronics",
						"Description": "13-inch MacBook Pro with M2 chip",
					},
					"Quantity": 1,
					"Subtotal": 1999.99,
				},
				{
					"Product": map[string]any{
						"ID":          102,
						"Name":        "Wireless Mouse",
						"Price":       79.99,
						"Category":    "Accessories",
						"Description": "Bluetooth wireless mouse",
					},
					"Quantity": 2,
					"Subtotal": 159.98,
				},
			},
			"Payment": map[string]any{
				"Method":      "credit_card",
				"Amount":      2159.97,
				"Currency":    "USD",
				"Transaction": "txn_123456789",
				"Status":      "completed",
			},
			"Shipping": map[string]any{
				"Address": map[string]any{
					"Street":     "123 Main St",
					"City":       "San Francisco",
					"State":      "CA",
					"PostalCode": "94105",
					"Country":    "USA",
				},
				"Method":      "express",
				"TrackingNum": "1Z999AA1234567890",
				"Estimated":   "2024-01-15",
			},
			"OrderStatus": map[string]any{
				"Status":    "shipped",
				"UpdatedAt": "2024-01-10T10:30:00Z",
				"Notes":     "Package is on the way",
			},
			"OrderID":   "ORD-2024-001",
			"Total":     2159.97,
			"CreatedAt": "2024-01-08T14:20:00Z",
		}

		var order Order
		err := decoder.Decode(input, &order)
		if err != nil {
			t.Fatalf("Order decoding failed: %v", err)
		}

		// Verify user info
		if order.User.ID != 1001 {
			t.Errorf("Expected User.ID 1001, got %d", order.User.ID)
		}
		if order.User.Username != "john_doe" {
			t.Errorf("Expected User.Username 'john_doe', got '%s'", order.User.Username)
		}
		if order.User.Email != "john@example.com" {
			t.Errorf("Expected User.Email 'john@example.com', got '%s'", order.User.Email)
		}

		// Verify order items
		if len(order.Items) != 2 {
			t.Errorf("Expected 2 items, got %d", len(order.Items))
		}
		if order.Items[0].Product.Name != "MacBook Pro" {
			t.Errorf("Expected first item name 'MacBook Pro', got '%s'", order.Items[0].Product.Name)
		}
		if order.Items[0].Quantity != 1 {
			t.Errorf("Expected first item quantity 1, got %d", order.Items[0].Quantity)
		}
		if order.Items[1].Product.Price != 79.99 {
			t.Errorf("Expected second item price 79.99, got %f", order.Items[1].Product.Price)
		}

		// Verify payment info
		if order.Payment.Method != "credit_card" {
			t.Errorf("Expected Payment.Method 'credit_card', got '%s'", order.Payment.Method)
		}
		if order.Payment.Amount != 2159.97 {
			t.Errorf("Expected Payment.Amount 2159.97, got %f", order.Payment.Amount)
		}
		if order.Payment.Status != "completed" {
			t.Errorf("Expected Payment.Status 'completed', got '%s'", order.Payment.Status)
		}

		// Verify shipping info
		if order.Shipping.Address.City != "San Francisco" {
			t.Errorf("Expected Shipping.Address.City 'San Francisco', got '%s'", order.Shipping.Address.City)
		}
		if order.Shipping.Method != "express" {
			t.Errorf("Expected Shipping.Method 'express', got '%s'", order.Shipping.Method)
		}
		if order.Shipping.TrackingNum != "1Z999AA1234567890" {
			t.Errorf("Expected Shipping.TrackingNum '1Z999AA1234567890', got '%s'", order.Shipping.TrackingNum)
		}

		// Verify order status
		if order.OrderStatus.Status != "shipped" {
			t.Errorf("Expected OrderStatus.Status 'shipped', got '%s'", order.OrderStatus.Status)
		}
		if order.OrderStatus.Notes != "Package is on the way" {
			t.Errorf("Expected OrderStatus.Notes 'Package is on the way', got '%s'", order.OrderStatus.Notes)
		}

		// Verify base order fields
		if order.OrderID != "ORD-2024-001" {
			t.Errorf("Expected OrderID 'ORD-2024-001', got '%s'", order.OrderID)
		}
		if order.Total != 2159.97 {
			t.Errorf("Expected Total 2159.97, got %f", order.Total)
		}
	})

	t.Run("Partial Order Decoding", func(t *testing.T) {
		input := map[string]any{
			"User": map[string]any{
				"ID":       2002,
				"Username": "jane_smith",
				"Email":    "jane@example.com",
			},
			"Items": []map[string]any{
				{
					"Product": map[string]any{
						"ID":    201,
						"Name":  "iPhone 15",
						"Price": 999.99,
					},
					"Quantity": 1,
					"Subtotal": 999.99,
				},
			},
			"Payment": map[string]any{
				"Method":   "paypal",
				"Amount":   999.99,
				"Currency": "USD",
				"Status":   "pending",
			},
			"OrderID": "ORD-2024-002",
			"Total":   999.99,
		}

		var order Order
		err := decoder.Decode(input, &order)
		if err != nil {
			t.Fatalf("Partial order decoding failed: %v", err)
		}

		// Verify present fields
		if order.User.ID != 2002 {
			t.Errorf("Expected User.ID 2002, got %d", order.User.ID)
		}
		if order.User.Username != "jane_smith" {
			t.Errorf("Expected User.Username 'jane_smith', got '%s'", order.User.Username)
		}
		if len(order.Items) != 1 {
			t.Errorf("Expected 1 item, got %d", len(order.Items))
		}
		if order.Items[0].Product.Name != "iPhone 15" {
			t.Errorf("Expected item name 'iPhone 15', got '%s'", order.Items[0].Product.Name)
		}

		// Verify zero values for missing fields
		if order.User.Phone != "" {
			t.Errorf("Expected User.Phone empty string, got '%s'", order.User.Phone)
		}
		if order.Shipping.Address.City != "" {
			t.Errorf("Expected Shipping.Address.City empty string, got '%s'", order.Shipping.Address.City)
		}
		if order.OrderStatus.Status != "" {
			t.Errorf("Expected OrderStatus.Status empty string, got '%s'", order.OrderStatus.Status)
		}
	})

	t.Run("Type Conversion in Order", func(t *testing.T) {
		input := map[string]any{
			"User": map[string]any{
				"ID":       "3003", // string to number
				"Username": "bob_wilson",
				"Email":    "bob@example.com",
			},
			"Items": []map[string]any{
				{
					"Product": map[string]any{
						"ID":    "301",
						"Name":  "Gaming Laptop",
						"Price": "1299.99", // string to float
					},
					"Quantity": "1", // string to int
					"Subtotal": "1299.99",
				},
			},
			"Payment": map[string]any{
				"Method": "debit_card",
				"Amount": "1299.99",
				"Status": "completed",
			},
			"OrderID": "ORD-2024-003",
			"Total":   "1299.99",
		}

		var order Order
		err := decoder.Decode(input, &order)
		if err != nil {
			t.Fatalf("Type conversion order failed: %v", err)
		}

		// Verify type conversion
		if order.User.ID != 3003 {
			t.Errorf("Expected User.ID 3003, got %d", order.User.ID)
		}
		if order.Items[0].Product.ID != 301 {
			t.Errorf("Expected Product.ID 301, got %d", order.Items[0].Product.ID)
		}
		if order.Items[0].Quantity != 1 {
			t.Errorf("Expected Quantity 1, got %d", order.Items[0].Quantity)
		}
		if order.Items[0].Product.Price != 1299.99 {
			t.Errorf("Expected Product.Price 1299.99, got %f", order.Items[0].Product.Price)
		}
		if order.Total != 1299.99 {
			t.Errorf("Expected Total 1299.99, got %f", order.Total)
		}
	})

	t.Run("Empty Order", func(t *testing.T) {
		input := map[string]any{}

		var order Order
		err := decoder.Decode(input, &order)
		if err != nil {
			t.Fatalf("Empty order decoding failed: %v", err)
		}

		// Verify all fields keep zero values
		if order.User.ID != 0 {
			t.Errorf("Expected User.ID 0, got %d", order.User.ID)
		}
		if order.User.Username != "" {
			t.Errorf("Expected User.Username empty string, got '%s'", order.User.Username)
		}
		if len(order.Items) != 0 {
			t.Errorf("Expected 0 items, got %d", len(order.Items))
		}
		if order.OrderID != "" {
			t.Errorf("Expected OrderID empty string, got '%s'", order.OrderID)
		}
		if order.Total != 0.0 {
			t.Errorf("Expected Total 0.0, got %f", order.Total)
		}
	})
}

// ===== Simpler scenario: user profile config =====

// User preferences
type UserPreferences struct {
	Theme         string
	Language      string
	Notifications bool
	Categories    []string
}

// User config (3-level nested structure)
type UserConfig struct {
	User        UserInfo
	Preferences UserPreferences
	LastLogin   time.Time
	IsActive    bool
}

// ===== Embedded field scenarios =====

// Base entity for embedding
type BaseEntity struct {
	ID        int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Auditable entity for embedding
type Auditable struct {
	CreatedBy string
	UpdatedBy string
}

// Article with embedded fields
type Article struct {
	BaseEntity // embedded
	Auditable  // embedded
	Title      string
	Content    string
	Published  bool
	ViewCount  int
}

// Multi-level embedded fields
type Timestamped struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Identifiable struct {
	Timestamped // embedded Timestamped
	ID          int
	UUID        string
}

type Author struct {
	Identifiable // embedded Identifiable (contains Timestamped)
	Name         string
	Email        string
}

func TestUserConfigDecoding(t *testing.T) {
	decoder := New()

	t.Run("Complete User Config", func(t *testing.T) {
		input := map[string]any{
			"User": map[string]any{
				"ID":       4001,
				"Username": "alice_johnson",
				"Email":    "alice@example.com",
				"Phone":    "+1-555-0456",
			},
			"Preferences": map[string]any{
				"Theme":         "dark",
				"Language":      "en-US",
				"Notifications": true,
				"Categories":    []string{"technology", "science", "business"},
			},
			"LastLogin": "2024-01-10T09:15:00Z",
			"IsActive":  true,
		}

		var config UserConfig
		err := decoder.Decode(input, &config)
		if err != nil {
			t.Fatalf("User config decoding failed: %v", err)
		}

		// Verify user info
		if config.User.ID != 4001 {
			t.Errorf("Expected User.ID 4001, got %d", config.User.ID)
		}
		if config.User.Username != "alice_johnson" {
			t.Errorf("Expected User.Username 'alice_johnson', got '%s'", config.User.Username)
		}

		// Verify preferences
		if config.Preferences.Theme != "dark" {
			t.Errorf("Expected Preferences.Theme 'dark', got '%s'", config.Preferences.Theme)
		}
		if config.Preferences.Language != "en-US" {
			t.Errorf("Expected Preferences.Language 'en-US', got '%s'", config.Preferences.Language)
		}
		if !config.Preferences.Notifications {
			t.Errorf("Expected Preferences.Notifications true, got %v", config.Preferences.Notifications)
		}
		if len(config.Preferences.Categories) != 3 {
			t.Errorf("Expected 3 categories, got %d", len(config.Preferences.Categories))
		}
		if config.Preferences.Categories[0] != "technology" {
			t.Errorf("Expected first category 'technology', got '%s'", config.Preferences.Categories[0])
		}

		// Verify config status
		if !config.IsActive {
			t.Errorf("Expected IsActive true, got %v", config.IsActive)
		}
	})
}

func TestAnonymousFieldDecoding(t *testing.T) {
	decoder := New()

	t.Run("Single Anonymous Field", func(t *testing.T) {
		input := map[string]any{
			// BaseEntity fields (embedded)
			"ID":        1001,
			"CreatedAt": "2024-01-01 10:00:00Z",
			"UpdatedAt": "2024-01-02 15:30:00Z",

			// Auditable fields (embedded)
			"CreatedBy": "admin",
			"UpdatedBy": "editor",

			// Article-owned fields
			"Title":     "Go best practices",
			"Content":   "This is an article about Go...",
			"Published": true,
			"ViewCount": 1500,
		}

		var article Article
		err := decoder.Decode(input, &article)
		if err != nil {
			t.Fatalf("Article decoding failed: %v", err)
		}

		// Verify embedded BaseEntity fields
		if article.ID != 1001 {
			t.Errorf("Expected ID 1001, got %d", article.ID)
		}
		if article.BaseEntity.ID != 1001 {
			t.Errorf("Expected BaseEntity.ID 1001, got %d", article.BaseEntity.ID)
		}

		// Verify time field conversion
		expectedCreatedAt, _ := time.Parse("2006-01-02 15:04:05Z07:00", "2024-01-01 10:00:00Z")
		if !article.CreatedAt.Equal(expectedCreatedAt) {
			t.Errorf("Expected CreatedAt %v, got %v", expectedCreatedAt, article.CreatedAt)
		}
		if !article.BaseEntity.CreatedAt.Equal(expectedCreatedAt) {
			t.Errorf("Expected BaseEntity.CreatedAt %v, got %v", expectedCreatedAt, article.BaseEntity.CreatedAt)
		}

		expectedUpdatedAt, _ := time.Parse("2006-01-02 15:04:05Z07:00", "2024-01-02 15:30:00Z")
		if !article.UpdatedAt.Equal(expectedUpdatedAt) {
			t.Errorf("Expected UpdatedAt %v, got %v", expectedUpdatedAt, article.UpdatedAt)
		}
		if !article.BaseEntity.UpdatedAt.Equal(expectedUpdatedAt) {
			t.Errorf("Expected BaseEntity.UpdatedAt %v, got %v", expectedUpdatedAt, article.BaseEntity.UpdatedAt)
		}

		// Verify embedded Auditable fields
		if article.CreatedBy != "admin" {
			t.Errorf("Expected CreatedBy 'admin', got '%s'", article.CreatedBy)
		}
		if article.UpdatedBy != "editor" {
			t.Errorf("Expected UpdatedBy 'editor', got '%s'", article.UpdatedBy)
		}

		// Verify article-owned fields
		if article.Title != "Go best practices" {
			t.Errorf("Expected Title 'Go best practices', got '%s'", article.Title)
		}
		if article.Content != "This is an article about Go..." {
			t.Errorf("Expected Content 'This is an article about Go...', got '%s'", article.Content)
		}
		if !article.Published {
			t.Errorf("Expected Published true, got %v", article.Published)
		}
		if article.ViewCount != 1500 {
			t.Errorf("Expected ViewCount 1500, got %d", article.ViewCount)
		}
	})

	t.Run("Nested Anonymous Fields", func(t *testing.T) {
		input := map[string]any{
			// Timestamped fields (via embedded Identifiable)
			"CreatedAt": "2024-01-01T08:00:00Z",
			"UpdatedAt": "2024-01-05T12:00:00Z",

			// Identifiable fields (embedded)
			"ID":   2001,
			"UUID": "550e8400-e29b-41d4-a716-446655440000",

			// Author-owned fields
			"Name":  "alice",
			"Email": "zhangsan@example.com",
		}

		var author Author
		err := decoder.Decode(input, &author)
		if err != nil {
			t.Fatalf("Author decoding failed: %v", err)
		}

		// Verify multi-level embedded fields
		if author.ID != 2001 {
			t.Errorf("Expected ID 2001, got %d", author.ID)
		}
		if author.UUID != "550e8400-e29b-41d4-a716-446655440000" {
			t.Errorf("Expected UUID '550e8400-e29b-41d4-a716-446655440000', got '%s'", author.UUID)
		}

		// Verify deepest embedded time conversion
		expectedCreatedAt, _ := time.Parse(time.RFC3339, "2024-01-01T08:00:00Z")
		if author.Timestamped.CreatedAt.IsZero() {
			t.Errorf("Expected CreatedAt to be set")
		}
		if !author.CreatedAt.Equal(expectedCreatedAt) {
			t.Errorf("Expected CreatedAt %v, got %v", expectedCreatedAt, author.CreatedAt)
		}
		if !author.Timestamped.CreatedAt.Equal(expectedCreatedAt) {
			t.Errorf("Expected Timestamped.CreatedAt %v, got %v", expectedCreatedAt, author.Timestamped.CreatedAt)
		}
		if !author.Identifiable.Timestamped.CreatedAt.Equal(expectedCreatedAt) {
			t.Errorf("Expected Identifiable.Timestamped.CreatedAt %v, got %v", expectedCreatedAt, author.Identifiable.Timestamped.CreatedAt)
		}

		expectedUpdatedAt, _ := time.Parse(time.RFC3339, "2024-01-05T12:00:00Z")
		if author.Identifiable.Timestamped.UpdatedAt.IsZero() {
			t.Errorf("Expected UpdatedAt to be set")
		}
		if !author.UpdatedAt.Equal(expectedUpdatedAt) {
			t.Errorf("Expected UpdatedAt %v, got %v", expectedUpdatedAt, author.UpdatedAt)
		}
		if !author.Timestamped.UpdatedAt.Equal(expectedUpdatedAt) {
			t.Errorf("Expected Timestamped.UpdatedAt %v, got %v", expectedUpdatedAt, author.Timestamped.UpdatedAt)
		}
		if !author.Identifiable.Timestamped.UpdatedAt.Equal(expectedUpdatedAt) {
			t.Errorf("Expected Identifiable.Timestamped.UpdatedAt %v, got %v", expectedUpdatedAt, author.Identifiable.Timestamped.UpdatedAt)
		}

		// Verify author-owned fields
		if author.Name != "alice" {
			t.Errorf("Expected Name 'alice', got '%s'", author.Name)
		}
		if author.Email != "zhangsan@example.com" {
			t.Errorf("Expected Email 'zhangsan@example.com', got '%s'", author.Email)
		}
	})

	t.Run("Partial Anonymous Fields", func(t *testing.T) {
		input := map[string]any{
			// Only provide partial fields
			"ID":        3001,
			"Title":     "partial fields test",
			"Published": false,
		}

		var article Article
		err := decoder.Decode(input, &article)
		if err != nil {
			t.Fatalf("Partial article decoding failed: %v", err)
		}

		// Verify present fields
		if article.ID != 3001 {
			t.Errorf("Expected ID 3001, got %d", article.ID)
		}
		if article.Title != "partial fields test" {
			t.Errorf("Expected Title 'partial fields test', got '%s'", article.Title)
		}
		if article.Published != false {
			t.Errorf("Expected Published false, got %v", article.Published)
		}

		// Verify zero values for missing fields
		if article.CreatedBy != "" {
			t.Errorf("Expected CreatedBy empty string, got '%s'", article.CreatedBy)
		}
		if article.Content != "" {
			t.Errorf("Expected Content empty string, got '%s'", article.Content)
		}
		if article.ViewCount != 0 {
			t.Errorf("Expected ViewCount 0, got %d", article.ViewCount)
		}

		// Verify zero value times
		if !article.CreatedAt.IsZero() {
			t.Errorf("Expected CreatedAt to be zero, got %v", article.CreatedAt)
		}
		if !article.UpdatedAt.IsZero() {
			t.Errorf("Expected UpdatedAt to be zero, got %v", article.UpdatedAt)
		}
		if !article.BaseEntity.CreatedAt.IsZero() {
			t.Errorf("Expected BaseEntity.CreatedAt to be zero, got %v", article.BaseEntity.CreatedAt)
		}
		if !article.BaseEntity.UpdatedAt.IsZero() {
			t.Errorf("Expected BaseEntity.UpdatedAt to be zero, got %v", article.BaseEntity.UpdatedAt)
		}
	})

	t.Run("Type Conversion with Anonymous Fields", func(t *testing.T) {
		input := map[string]any{
			// Use numeric strings
			"ID":        "4001",
			"ViewCount": "999",
			"Title":     "type conversion test",
			"Published": "true",
			"CreatedBy": "system",
			// Use different time formats
			"CreatedAt": "2024-03-15",      // date-only format
			"UpdatedAt": int64(1710518400), // Unix timestamp (2024-03-15 12:00:00 UTC)
		}

		var article Article
		err := decoder.Decode(input, &article)
		if err != nil {
			t.Fatalf("Type conversion with anonymous fields failed: %v", err)
		}

		// Verify type conversion
		if article.ID != 4001 {
			t.Errorf("Expected ID 4001, got %d", article.ID)
		}
		if article.ViewCount != 999 {
			t.Errorf("Expected ViewCount 999, got %d", article.ViewCount)
		}
		if !article.Published {
			t.Errorf("Expected Published true, got %v", article.Published)
		}

		// Verify time conversion
		expectedCreatedAt, _ := time.Parse("2006-01-02", "2024-03-15")
		if !article.CreatedAt.Equal(expectedCreatedAt) {
			t.Errorf("Expected CreatedAt %v, got %v", expectedCreatedAt, article.CreatedAt)
		}

		expectedUpdatedAt := time.Unix(1710518400, 0)
		if !article.UpdatedAt.Equal(expectedUpdatedAt) {
			t.Errorf("Expected UpdatedAt %v, got %v", expectedUpdatedAt, article.UpdatedAt)
		}
	})
}

// ===== Benchmarks =====

func BenchmarkOrderDecoding(b *testing.B) {
	decoder := New()

	input := map[string]any{
		"User": map[string]any{
			"ID":       1001,
			"Username": "test_user",
			"Email":    "test@example.com",
			"Phone":    "+1-555-0000",
		},
		"Items": []map[string]any{
			{
				"Product": map[string]any{
					"ID":    101,
					"Name":  "Test Product",
					"Price": 99.99,
				},
				"Quantity": 1,
				"Subtotal": 99.99,
			},
		},
		"Payment": map[string]any{
			"Method": "credit_card",
			"Amount": 99.99,
			"Status": "completed",
		},
		"OrderID": "BENCH-001",
		"Total":   99.99,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var order Order
		decoder.Decode(input, &order)
	}
}

// ===== Duration decoding tests =====

type DurationConfig struct {
	Timeout     time.Duration `yaml:"timeout"`
	Interval    time.Duration `yaml:"interval"`
	MaxWaitTime time.Duration `yaml:"maxWaitTime"`
}

func TestDurationDecoding(t *testing.T) {
	decoder := New()

	t.Run("String format duration", func(t *testing.T) {
		input := map[string]any{
			"timeout":     "30s",
			"interval":    "5m",
			"maxWaitTime": "1h",
		}

		var config DurationConfig
		err := decoder.Decode(input, &config)
		if err != nil {
			t.Fatalf("Duration decoding failed: %v", err)
		}

		if config.Timeout != 30*time.Second {
			t.Errorf("Expected timeout 30s, got %v", config.Timeout)
		}
		if config.Interval != 5*time.Minute {
			t.Errorf("Expected interval 5m, got %v", config.Interval)
		}
		if config.MaxWaitTime != 1*time.Hour {
			t.Errorf("Expected maxWaitTime 1h, got %v", config.MaxWaitTime)
		}
	})

	t.Run("Int64 as seconds", func(t *testing.T) {
		input := map[string]any{
			"timeout":     int64(30),   // 30 seconds
			"interval":    int64(300),  // 300 seconds = 5 minutes
			"maxWaitTime": int64(3600), // 3600 seconds = 1 hour
		}

		var config DurationConfig
		err := decoder.Decode(input, &config)
		if err != nil {
			t.Fatalf("Duration decoding failed: %v", err)
		}

		if config.Timeout != 30*time.Second {
			t.Errorf("Expected timeout 30s, got %v", config.Timeout)
		}
		if config.Interval != 300*time.Second {
			t.Errorf("Expected interval 300s, got %v", config.Interval)
		}
		if config.MaxWaitTime != 3600*time.Second {
			t.Errorf("Expected maxWaitTime 3600s, got %v", config.MaxWaitTime)
		}
	})

	t.Run("Int as seconds", func(t *testing.T) {
		input := map[string]any{
			"timeout":  15, // 15 seconds
			"interval": 60, // 60 seconds
		}

		var config DurationConfig
		err := decoder.Decode(input, &config)
		if err != nil {
			t.Fatalf("Duration decoding failed: %v", err)
		}

		if config.Timeout != 15*time.Second {
			t.Errorf("Expected timeout 15s, got %v", config.Timeout)
		}
		if config.Interval != 60*time.Second {
			t.Errorf("Expected interval 60s, got %v", config.Interval)
		}
	})

	t.Run("Float64 as seconds", func(t *testing.T) {
		input := map[string]any{
			"timeout":     30.5, // 30.5 seconds
			"interval":    1.5,  // 1.5 seconds
			"maxWaitTime": 0.5,  // 0.5 seconds
		}

		var config DurationConfig
		err := decoder.Decode(input, &config)
		if err != nil {
			t.Fatalf("Duration decoding failed: %v", err)
		}

		expectedTimeout := time.Duration(30.5 * float64(time.Second))
		if config.Timeout != expectedTimeout {
			t.Errorf("Expected timeout %v, got %v", expectedTimeout, config.Timeout)
		}

		expectedInterval := time.Duration(1.5 * float64(time.Second))
		if config.Interval != expectedInterval {
			t.Errorf("Expected interval %v, got %v", expectedInterval, config.Interval)
		}

		expectedMaxWaitTime := time.Duration(0.5 * float64(time.Second))
		if config.MaxWaitTime != expectedMaxWaitTime {
			t.Errorf("Expected maxWaitTime %v, got %v", expectedMaxWaitTime, config.MaxWaitTime)
		}
	})

	t.Run("Consistent behavior: int64 and float64 both as seconds", func(t *testing.T) {
		// Verify int64 and float64 are both interpreted as seconds
		input1 := map[string]any{
			"timeout": int64(30),
		}
		input2 := map[string]any{
			"timeout": 30.0,
		}

		var config1, config2 DurationConfig
		err1 := decoder.Decode(input1, &config1)
		err2 := decoder.Decode(input2, &config2)
		if err1 != nil || err2 != nil {
			t.Fatalf("Duration decoding failed: err1=%v, err2=%v", err1, err2)
		}

		// Both should produce the same result within minor float error
		diff := config1.Timeout - config2.Timeout
		if diff < 0 {
			diff = -diff
		}
		if diff > time.Millisecond {
			t.Errorf("Expected consistent behavior: int64(30) and 30.0 should produce similar results, got %v vs %v", config1.Timeout, config2.Timeout)
		}
	})
}

type mapDBConfig struct {
	DSN          string `yaml:"dsn"`
	MaxOpenConns int    `yaml:"maxOpenConns"`
}

type mapRedisConfig struct {
	Addr string `yaml:"addr"`
	DB   int    `yaml:"db"`
}

type mapResourceConfig struct {
	MySQL map[string]mapDBConfig     `yaml:"mysql"`
	Redis map[string]*mapRedisConfig `yaml:"redis"`
}

func TestMapStringStructDecoding(t *testing.T) {
	decoder := New()

	input := map[string]any{
		"mysql": map[string]any{
			"primary": map[string]any{
				"dsn":          "root:pass@tcp(localhost:3306)/app",
				"maxOpenConns": 100,
			},
			"readonly": map[string]any{
				"dsn":          "root:pass@tcp(localhost:3306)/app_read",
				"maxOpenConns": 20,
			},
		},
		"redis": map[string]any{
			"cache": map[string]any{
				"addr": "127.0.0.1:6379",
				"db":   0,
			},
		},
	}

	var cfg mapResourceConfig
	if err := decoder.Decode(input, &cfg); err != nil {
		t.Fatalf("map decode failed: %v", err)
	}

	if len(cfg.MySQL) != 2 {
		t.Fatalf("expected 2 mysql configs, got %d", len(cfg.MySQL))
	}
	if cfg.MySQL["primary"].DSN == "" {
		t.Fatalf("expected mysql.primary.dsn to be decoded")
	}
	if cfg.MySQL["readonly"].MaxOpenConns != 20 {
		t.Fatalf("expected mysql.readonly.maxOpenConns=20, got %d", cfg.MySQL["readonly"].MaxOpenConns)
	}
	if cfg.Redis["cache"] == nil {
		t.Fatalf("expected redis.cache pointer to be decoded")
	}
	if cfg.Redis["cache"].Addr != "127.0.0.1:6379" {
		t.Fatalf("unexpected redis.cache.addr: %s", cfg.Redis["cache"].Addr)
	}
}

func TestDecode_DoublePointerTarget(t *testing.T) {
	decoder := New()
	input := map[string]any{
		"timeout": "30s",
	}

	var cfg *DurationConfig
	if err := decoder.Decode(input, &cfg); err != nil {
		t.Fatalf("decode to **T failed: %v", err)
	}
	if cfg == nil {
		t.Fatalf("expected cfg to be allocated")
	}
	if cfg.Timeout != 30*time.Second {
		t.Fatalf("expected timeout 30s, got %v", cfg.Timeout)
	}
}
