package main

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"type:text;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	SKU         string         `gorm:"type:varchar(128)" json:"sku"`
	Price       float64        `gorm:"type:decimal(10,2);not null" json:"price"`
	Quantity    int            `gorm:"type:int;not null" json:"quantity"`
	Category    string         `gorm:"type:text" json:"category"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type ProductCreateRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description,omitempty"`
	SKU         string  `json:"sku" validate:"required"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Quantity    int     `json:"quantity" validate:"min=0"`
	Category    string  `json:"category,omitempty"`
}

type ProductUpdateRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	SKU         *string  `json:"sku,omitempty"`
	Price       *float64 `json:"price,omitempty" validate:"omitempty,gt=0"`
	Quantity    *int     `json:"quantity,omitempty" validate:"omitempty,min=0"`
	Category    *string  `json:"category,omitempty"`
}

type BulkProductResponse struct {
	Products   []Product `json:"products"`
	Page       int       `json:"page"`
	Size       int       `json:"size"`
	TotalPages int64     `json:"total_pages"`
	TotalCount int64     `json:"total_count"`
}
