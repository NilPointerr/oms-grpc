package models

import (
	"time"

	"gorm.io/gorm"
)

// Order represents an order in the OMS system
type Order struct {
	ID         int32            `json:"id"`
	UserID     int32            `json:"user_id"`
	TotalPrice float64        `json:"total_price"`
	Status     string         `json:"status"`
	FinalPrice float64        `json:"final_price"` // Total price after applying discounts
	Items      []OrderItem    `json:"items"`       // List of items in the order
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID        int32            `json:"id"`
	OrderID   int32            `json:"order_id"`
	ItemID    int32            `json:"item_id"`
	Quantity  int32           `json:"quantity"`
	Price     float64        `json:"price"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

// type OrderResposnse struct {
// 	ID         int32                 `json:"id"`
// 	UserID     int32                 `json:"user_id"`
// 	TotalPrice float64             `json:"total_price"`
// 	Status     string              `json:"status"`
// 	FinalPrice float64             `json:"final_price"` // Total price after applying discounts
// 	Items      []ResponseOrderItem `json:"items"`       // List of items in the order
// 	CreatedAt  time.Time           `json:"created_at"`
// 	UpdatedAt  time.Time           `json:"updated_at"`
// 	DeletedAt  gorm.DeletedAt      `json:"deleted_at"`
// }

// type ResponseOrderItem struct {
// 	ItemID int `json:"item_id"`
// 	// ItemName string `json:"item_name"`
// 	Quantity int     `json:"quantity"`
// 	Price    float64 `json:"price"`
// }

// type OrderResposnseGet struct {
// 	ID         int32                    `json:"id"`
// 	UserID     int32                    `json:"user_id"`
// 	TotalPrice float64                `json:"total_price"`
// 	Status     string                 `json:"status"`
// 	FinalPrice float64                `json:"final_price"` // Total price after applying discounts
// 	Items      []ResponseOrderItemGet `json:"items"`       // List of items in the order
// 	CreatedAt  time.Time              `json:"created_at"`
// 	UpdatedAt  time.Time              `json:"updated_at"`
// 	DeletedAt  gorm.DeletedAt         `json:"deleted_at"`
// }

// type ResponseOrderItemGet struct {
// 	ItemID   int32     `json:"item_id"`
// 	ItemName string  `json:"item_name"`
// 	Quantity int32     `json:"quantity"`
// 	Price    float64 `json:"price"`
// }
