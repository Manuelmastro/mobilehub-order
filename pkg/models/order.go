package models

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	//ID        uint        `gorm:"primaryKey"`        // Unique ID for the order
	UserID  string      `gorm:"not null"`           // ID of the user placing the order
	Address string      `gorm:"not null"`           // Shipping address for the order
	Items   []OrderItem `gorm:"foreignKey:OrderID"` // Associated order items
	Total   float64     `gorm:"not null"`           // Total price of the order
	Payment string      `gorm:"not null"`           // Payment method (e.g., COD, online)
	Status  string      `gorm:"not null"`           // Status of the order (e.g., Pending, Completed)
	//CreatedAt time.Time   // Timestamp for when the order was created
	//UpdatedAt time.Time   // Timestamp for when the order was last updated
}

type OrderItem struct {
	gorm.Model
	//ID          uint    `gorm:"primaryKey"` // Unique ID for the order item
	OrderID     uint    `gorm:"not null"` // Foreign key linking to the order
	ProductID   string  `gorm:"not null"` // ID of the product
	ProductName string  `gorm:"not null"` // Name of the product
	Price       float64 `gorm:"not null"` // Price of a single unit of the product
	Quantity    int     `gorm:"not null"` // Quantity of the product in the order
	TotalPrice  float64 `gorm:"not null"` // Total price (Price * Quantity)
}
