package model

import (
	"time"
)

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	ID         int64       `gorm:"primaryKey;autoIncrement"`
	Number     string      `gorm:"uniqueIndex;not null"`
	UserID     int64       `gorm:"index;not null"`
	User       User        `gorm:"foreignKey:UserID;references:ID"`
	Status     OrderStatus `gorm:"not null"`
	Accrual    *float64
	UploadedAt time.Time `gorm:"autoCreateTime"`
}
