package main

import "time"

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Address struct {
	Street  string `json:"street" dynamodbav:"street"`
	State   string `json:"state,omitempty" dynamodbav:"state,omitempty"`
	Country string `json:"country" dynamodbav:"country"`
}

type User struct {
	Username  string             `json:"username" dynamodbav:"-"`
	FullName  string             `json:"full_name,omitempty" dynamodbav:"full_name,omitempty"`
	Email     string             `json:"email,omitempty" dynamodbav:"email,omitempty"`
	Addresses map[string]Address `json:"addresses,omitempty" dynamodbav:"addresses,omitempty"`
}

type Order struct {
	ID         string      `json:"id" dynamodbav:"order_id"`
	UserID     string      `json:"user_id" dynamodbav:"user_id"`
	Status     OrderStatus `json:"status" dynamodbav:"status"`
	AddressKey string      `json:"address_key" dynamodbav:"address_key"`
	CreatedAt  time.Time   `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at" dynamodbav:"updated_at"`
}

type OrderItem struct {
	OrderID     string  `json:"order_id" dynamodbav:"order_id"`
	ItemID      string  `json:"item_id" dynamodbav:"item_id"`
	Name        string  `json:"name" dynamodbav:"name"`
	Description string  `json:"description" dynamodbav:"description"`
	Price       float64 `json:"price" dynamodbav:"price"`
	Quantity    int     `json:"quantity" dynamodbav:"quantity"`
}
