package types

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID         uuid.UUID `db:"id" json:"id,omitempty"`
	Name       string    `db:"name" json:"name,omitempty"`
	Img        *string   `db:"img" json:"img,omitempty"`
	Email      string    `db:"email" json:"email,omitempty"`
	Phone      string    `db:"phone" json:"phone,omitempty"`
	Password   string    `db:"password" json:"-"`
	Created_at string    `db:"created_at" json:"created_at,omitempty"`
	Updated_at string    `db:"updated_at" json:"updated_at,omitempty"`
	Roles      []int     `db:"roles" json:"roles,omitempty"`
}

type Vendor struct {
	ID          uuid.UUID `db:"id"          json:"id,omitempty"`
	Name        string    `db:"name"        json:"name,omitempty"`
	Img         *string   `db:"img"         json:"img,omitempty"`
	Description string    `db:"description" json:"description,omitempty"`
	Created_at  time.Time `db:"created_at"  json:"created_at,omitempty"`
	Updated_at  time.Time `db:"updated_at"  json:"updated_at,omitempty"`
}

type Role struct {
	ID   int    `db:"id" json:"id,omitempty"`
	Name string `db:"name" json:"name,omitempty"`
}

type Item struct {
	ID         uuid.UUID `db:"id"          json:"id,omitempty"`
	VendorId   uuid.UUID `db:"vendor_id"   json:"vendor_id,omitempty"`
	Name       string    `db:"name"        json:"name,omitempty"`
	Price      float64   `db:"price"       json:"price,omitempty"`
	Img        *string   `db:"img"         json:"img,omitempty"`
	Created_at time.Time `db:"created_at"  json:"created_at,omitempty"`
	Updated_at time.Time `db:"updated_at"  json:"updated_at,omitempty"`
}

type Order struct {
	ID             uuid.UUID    `db:"id"          json:"id,omitempty"`
	TotalOrderCost float64      `db:"total_order_cost" json:"total_order_cost,omitempty"`
	VendorId       uuid.UUID    `db:"vendor_id"   json:"vendor_id,omitempty"`
	CustomerId     uuid.UUID    `db:"customer_id"   json:"customer_id,omitempty"`
	Status         string       `db:"status"        json:"status,omitempty"`
	Created_at     time.Time    `db:"created_at"  json:"created_at,omitempty"`
	Updated_at     time.Time    `db:"updated_at"  json:"updated_at,omitempty"`
	OrderItems     []OrderItems `db:"-" json:"order_items,omitempty"`
}

type OrderItems struct {
	ID       uuid.UUID `db:"id"          json:"id,omitempty"`
	OrderId  uuid.UUID `db:"order_id"     json:"order_id,omitempty"`
	Quantity int       `db:"quantity"    json:"quantity,omitempty"`
	Price    float64   `db:"price"       json:"price,omitempty"`
	ItemId   uuid.UUID `db:"item_id"     json:"item_id,omitempty"`
}

type Cart struct {
	ID         uuid.UUID   `db:"id"          json:"id,omitempty"`
	TotalPrice float64     `db:"total_price" json:"total_price,omitempty"`
	Quantity   int         `db:"quantity"    json:"quantity,omitempty"`
	VendorId   uuid.UUID   `db:"vendor_id"   json:"vendor_id,omitempty"`
	Created_at time.Time   `db:"created_at"  json:"created_at,omitempty"`
	Updated_at time.Time   `db:"updated_at"  json:"updated_at,omitempty"`
	CartItem   []CartItems `db:"-" json:"cart_item,omitempty"`
}

type CartItems struct {
	CartId   uuid.UUID `db:"cart_id"     json:"cart_id,omitempty"`
	Quantity int       `db:"quantity"    json:"quantity,omitempty"`
	ItemId   uuid.UUID `db:"item_id"     json:"item_id,omitempty"`
}

type Table struct {
	ID           uuid.UUID `db:"id"          json:"id,omitempty"`
	Name         string    `db:"name"        json:"name,omitempty"`
	VendorId     uuid.UUID `db:"vendor_id"   json:"vendor_id,omitempty"`
	CustomerId   uuid.UUID `db:"customer_id"   json:"customer_id,omitempty"`
	IsAvailable  bool      `db:"is_available"        json:"is_available,omitempty"`
	NeedsService bool      `db:"is_needs_service" json:"needs_service,omitempty"`
}

type Meta struct {
	Total       int `json:"total,omitempty"`
	PerPage     int `json:"per_page,omitempty"`
	CurrentPage int `json:"current_page,omitempty"`
	FirstPage   int `json:"first_page,omitempty"`
	LastPage    int `json:"last_page,omitempty"`
	From        int `json:"from,omitempty"`
	To          int `json:"to,omitempty"`
}

type Response struct {
	Meta interface{} `json:"meta,omitempty"`
	Data interface{} `json:"data,omitempty"`
}
