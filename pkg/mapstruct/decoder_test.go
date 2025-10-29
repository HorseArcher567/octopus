package mapstruct

import (
	"testing"
	"time"
)

// ===== 业务场景：电商系统用户订单 =====

// 用户基础信息
type UserInfo struct {
	ID       int
	Username string
	Email    string
	Phone    string
}

// 地址信息
type Address struct {
	Street     string
	City       string
	State      string
	PostalCode string
	Country    string
}

// 商品信息
type Product struct {
	ID          int
	Name        string
	Price       float64
	Category    string
	Description string
}

// 订单项
type OrderItem struct {
	Product  Product
	Quantity int
	Subtotal float64
}

// 支付信息
type PaymentInfo struct {
	Method      string
	Amount      float64
	Currency    string
	Transaction string
	Status      string
}

// 配送信息
type ShippingInfo struct {
	Address     Address
	Method      string
	TrackingNum string
	Estimated   string
}

// 订单状态
type OrderStatus struct {
	Status    string
	UpdatedAt time.Time
	Notes     string
}

// 完整订单（5层嵌套结构）
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

// ===== 测试用例 =====

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

		// 验证用户信息
		if order.User.ID != 1001 {
			t.Errorf("Expected User.ID 1001, got %d", order.User.ID)
		}
		if order.User.Username != "john_doe" {
			t.Errorf("Expected User.Username 'john_doe', got '%s'", order.User.Username)
		}
		if order.User.Email != "john@example.com" {
			t.Errorf("Expected User.Email 'john@example.com', got '%s'", order.User.Email)
		}

		// 验证订单项
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

		// 验证支付信息
		if order.Payment.Method != "credit_card" {
			t.Errorf("Expected Payment.Method 'credit_card', got '%s'", order.Payment.Method)
		}
		if order.Payment.Amount != 2159.97 {
			t.Errorf("Expected Payment.Amount 2159.97, got %f", order.Payment.Amount)
		}
		if order.Payment.Status != "completed" {
			t.Errorf("Expected Payment.Status 'completed', got '%s'", order.Payment.Status)
		}

		// 验证配送信息
		if order.Shipping.Address.City != "San Francisco" {
			t.Errorf("Expected Shipping.Address.City 'San Francisco', got '%s'", order.Shipping.Address.City)
		}
		if order.Shipping.Method != "express" {
			t.Errorf("Expected Shipping.Method 'express', got '%s'", order.Shipping.Method)
		}
		if order.Shipping.TrackingNum != "1Z999AA1234567890" {
			t.Errorf("Expected Shipping.TrackingNum '1Z999AA1234567890', got '%s'", order.Shipping.TrackingNum)
		}

		// 验证订单状态
		if order.OrderStatus.Status != "shipped" {
			t.Errorf("Expected OrderStatus.Status 'shipped', got '%s'", order.OrderStatus.Status)
		}
		if order.OrderStatus.Notes != "Package is on the way" {
			t.Errorf("Expected OrderStatus.Notes 'Package is on the way', got '%s'", order.OrderStatus.Notes)
		}

		// 验证订单基本信息
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

		// 验证存在的字段
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

		// 验证缺失字段的零值
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
				"ID":       "3003", // 字符串转数字
				"Username": "bob_wilson",
				"Email":    "bob@example.com",
			},
			"Items": []map[string]any{
				{
					"Product": map[string]any{
						"ID":    "301",
						"Name":  "Gaming Laptop",
						"Price": "1299.99", // 字符串转浮点数
					},
					"Quantity": "1", // 字符串转整数
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

		// 验证类型转换
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

		// 验证所有字段都是零值
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

// ===== 简化场景：用户配置文件 =====

// 用户偏好设置
type UserPreferences struct {
	Theme         string
	Language      string
	Notifications bool
	Categories    []string
}

// 用户配置（3层嵌套）
type UserConfig struct {
	User        UserInfo
	Preferences UserPreferences
	LastLogin   time.Time
	IsActive    bool
}

// ===== 匿名成员测试场景 =====

// 基础实体（用于内嵌）
type BaseEntity struct {
	ID        int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// 可审计实体（用于内嵌）
type Auditable struct {
	CreatedBy string
	UpdatedBy string
}

// 文章（包含匿名成员）
type Article struct {
	BaseEntity // 匿名内嵌
	Auditable  // 匿名内嵌
	Title      string
	Content    string
	Published  bool
	ViewCount  int
}

// 多层匿名内嵌
type Timestamped struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Identifiable struct {
	Timestamped // 匿名内嵌 Timestamped
	ID          int
	UUID        string
}

type Author struct {
	Identifiable // 匿名内嵌 Identifiable（包含 Timestamped）
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

		// 验证用户信息
		if config.User.ID != 4001 {
			t.Errorf("Expected User.ID 4001, got %d", config.User.ID)
		}
		if config.User.Username != "alice_johnson" {
			t.Errorf("Expected User.Username 'alice_johnson', got '%s'", config.User.Username)
		}

		// 验证偏好设置
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

		// 验证配置状态
		if !config.IsActive {
			t.Errorf("Expected IsActive true, got %v", config.IsActive)
		}
	})
}

func TestAnonymousFieldDecoding(t *testing.T) {
	decoder := New()

	t.Run("Single Anonymous Field", func(t *testing.T) {
		input := map[string]any{
			// BaseEntity 字段（匿名内嵌）
			"ID":        1001,
			"CreatedAt": "2024-01-01 10:00:00Z",
			"UpdatedAt": "2024-01-02 15:30:00Z",

			// Auditable 字段（匿名内嵌）
			"CreatedBy": "admin",
			"UpdatedBy": "editor",

			// Article 自有字段
			"Title":     "Go语言最佳实践",
			"Content":   "这是一篇关于Go语言的文章...",
			"Published": true,
			"ViewCount": 1500,
		}

		var article Article
		err := decoder.Decode(input, &article)
		if err != nil {
			t.Fatalf("Article decoding failed: %v", err)
		}

		// 验证 BaseEntity 匿名字段
		if article.ID != 1001 {
			t.Errorf("Expected ID 1001, got %d", article.ID)
		}
		if article.BaseEntity.ID != 1001 {
			t.Errorf("Expected BaseEntity.ID 1001, got %d", article.BaseEntity.ID)
		}

		// 验证时间字段转换是否正确
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

		// 验证 Auditable 匿名字段
		if article.CreatedBy != "admin" {
			t.Errorf("Expected CreatedBy 'admin', got '%s'", article.CreatedBy)
		}
		if article.UpdatedBy != "editor" {
			t.Errorf("Expected UpdatedBy 'editor', got '%s'", article.UpdatedBy)
		}

		// 验证 Article 自有字段
		if article.Title != "Go语言最佳实践" {
			t.Errorf("Expected Title 'Go语言最佳实践', got '%s'", article.Title)
		}
		if article.Content != "这是一篇关于Go语言的文章..." {
			t.Errorf("Expected Content '这是一篇关于Go语言的文章...', got '%s'", article.Content)
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
			// Timestamped 字段（通过 Identifiable 匿名内嵌）
			"CreatedAt": "2024-01-01T08:00:00Z",
			"UpdatedAt": "2024-01-05T12:00:00Z",

			// Identifiable 字段（匿名内嵌）
			"ID":   2001,
			"UUID": "550e8400-e29b-41d4-a716-446655440000",

			// Author 自有字段
			"Name":  "张三",
			"Email": "zhangsan@example.com",
		}

		var author Author
		err := decoder.Decode(input, &author)
		if err != nil {
			t.Fatalf("Author decoding failed: %v", err)
		}

		// 验证多层匿名内嵌字段
		if author.ID != 2001 {
			t.Errorf("Expected ID 2001, got %d", author.ID)
		}
		if author.UUID != "550e8400-e29b-41d4-a716-446655440000" {
			t.Errorf("Expected UUID '550e8400-e29b-41d4-a716-446655440000', got '%s'", author.UUID)
		}

		// 验证最深层的匿名内嵌字段 - 时间转换是否正确
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

		// 验证 Author 自有字段
		if author.Name != "张三" {
			t.Errorf("Expected Name '张三', got '%s'", author.Name)
		}
		if author.Email != "zhangsan@example.com" {
			t.Errorf("Expected Email 'zhangsan@example.com', got '%s'", author.Email)
		}
	})

	t.Run("Partial Anonymous Fields", func(t *testing.T) {
		input := map[string]any{
			// 只提供部分字段
			"ID":        3001,
			"Title":     "部分字段测试",
			"Published": false,
		}

		var article Article
		err := decoder.Decode(input, &article)
		if err != nil {
			t.Fatalf("Partial article decoding failed: %v", err)
		}

		// 验证存在的字段
		if article.ID != 3001 {
			t.Errorf("Expected ID 3001, got %d", article.ID)
		}
		if article.Title != "部分字段测试" {
			t.Errorf("Expected Title '部分字段测试', got '%s'", article.Title)
		}
		if article.Published != false {
			t.Errorf("Expected Published false, got %v", article.Published)
		}

		// 验证缺失字段的零值
		if article.CreatedBy != "" {
			t.Errorf("Expected CreatedBy empty string, got '%s'", article.CreatedBy)
		}
		if article.Content != "" {
			t.Errorf("Expected Content empty string, got '%s'", article.Content)
		}
		if article.ViewCount != 0 {
			t.Errorf("Expected ViewCount 0, got %d", article.ViewCount)
		}

		// 验证时间字段为零值
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
			// 使用字符串形式的数字
			"ID":        "4001",
			"ViewCount": "999",
			"Title":     "类型转换测试",
			"Published": "true",
			"CreatedBy": "system",
			// 使用不同的时间格式测试
			"CreatedAt": "2024-03-15",      // 日期格式
			"UpdatedAt": int64(1710518400), // Unix 时间戳 (2024-03-15 12:00:00 UTC)
		}

		var article Article
		err := decoder.Decode(input, &article)
		if err != nil {
			t.Fatalf("Type conversion with anonymous fields failed: %v", err)
		}

		// 验证类型转换
		if article.ID != 4001 {
			t.Errorf("Expected ID 4001, got %d", article.ID)
		}
		if article.ViewCount != 999 {
			t.Errorf("Expected ViewCount 999, got %d", article.ViewCount)
		}
		if !article.Published {
			t.Errorf("Expected Published true, got %v", article.Published)
		}

		// 验证时间类型转换
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

// ===== 性能测试 =====

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
