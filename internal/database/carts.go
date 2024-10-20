package database

import (
	"database/sql"

	"fmt"
	"net/http"
	"restaurant-management-backend/internal/helpers"
	"restaurant-management-backend/internal/types"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var cart_columns = []string{
	"id", "total_price", "quantity", "vendor_id", "created_at", "updated_at",
}

func (s *service) GetCart(userID string) (types.Cart, error) {
	var cart types.Cart
	query, args, err := QB.Select(strings.Join(cart_columns, ", ")).
		From("carts").
		Where("id = ?", userID).
		ToSql()
	if err != nil {
		return cart, err
	}
	err = s.db.Get(&cart, query, args...)
	return cart, err
}

func (s *service) GetCartItems(cartID uuid.UUID) ([]types.CartItems, error) {
	var cartItems []types.CartItems
	query, args, err := QB.Select("*").
		From("cart_items").
		Where("cart_id = ?", cartID).
		ToSql()
	if err != nil {
		return nil, err
	}
	err = s.db.Select(&cartItems, query, args...)
	return cartItems, err
}

var item_columns = []string{
	"id",
	"vendor_id",
	"name",
	"price",
	"created_at",
	"updated_at",
	helpers.ImageFormat,
}

func (s *service) GetCartItem(itemID uuid.UUID) (types.Item, error) {
	var item types.Item
	query, args, err := QB.Select(strings.Join(item_columns, ", ")).
		From("items").
		Where("id = ?", itemID).
		ToSql()
	if err != nil {
		return item, err
	}
	err = s.db.Get(&item, query, args...)
	return item, err
}

func (s *service) GetOrCreateCart(userID string, vendorID uuid.UUID) (types.Cart, error) {
	cart, err := s.GetCart(userID)
	if err == sql.ErrNoRows {
		return s.CreateCart(userID, vendorID)
	}
	if err != nil {
		return cart, err
	}
	if cart.VendorId != vendorID {
		return s.ResetCart(cart.ID, vendorID)
	}
	return cart, nil
}

func (s *service) CreateCart(userID string, vendorID uuid.UUID) (types.Cart, error) {
	cart := types.Cart{
		ID:         uuid.MustParse(userID),
		TotalPrice: 0,
		Quantity:   0,
		VendorId:   vendorID,
		Created_at: time.Now(),
		Updated_at: time.Now(),
	}
	query, args, err := QB.Insert("carts").
		Columns("id", "total_price", "quantity", "vendor_id", "created_at", "updated_at").
		Values(cart.ID, cart.TotalPrice, cart.Quantity, cart.VendorId, cart.Created_at, cart.Updated_at).
		ToSql()
	if err != nil {
		return cart, err
	}
	_, err = s.db.Exec(query, args...)
	return cart, err
}

func (s *service) ResetCart(cartID uuid.UUID, vendorID uuid.UUID) (types.Cart, error) {
	if err := s.ClearCartItems(cartID); err != nil {
		return types.Cart{}, err
	}
	query, args, err := QB.Update("carts").
		Set("vendor_id", vendorID).
		Set("total_price", 0).
		Set("quantity", 0).
		Set("updated_at", time.Now()).
		Where("id = ?", cartID).
		ToSql()
	if err != nil {
		return types.Cart{}, err
	}
	_, err = s.db.Exec(query, args...)
	if err != nil {
		return types.Cart{}, err
	}
	return s.GetCart(cartID.String())
}

func (s *service) ClearCartItems(cartID uuid.UUID) error {
	query, args, err := QB.Delete("cart_items").Where("cart_id = ?", cartID).ToSql()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(query, args...)
	return err
}

func (s *service) UpdateCartItem(cartID, itemID uuid.UUID, quantity int) error {
	var cartItem types.CartItems
	query, args, err := QB.Select("*").From("cart_items").
		Where("cart_id = ? AND item_id = ?", cartID, itemID).
		ToSql()
	if err != nil {
		return err
	}
	err = s.db.Get(&cartItem, query, args...)

	if err == sql.ErrNoRows {
		query, args, err = QB.Insert("cart_items").
			Columns("cart_id", "item_id", "quantity").
			Values(cartID, itemID, quantity).
			ToSql()
	} else if err == nil {
		query, args, err = QB.Update("cart_items").
			Set("quantity", quantity).
			Where("cart_id = ? AND item_id = ?", cartID, itemID).
			ToSql()
	} else {
		return err
	}

	if err != nil {
		return err
	}
	_, err = s.db.Exec(query, args...)
	return err
}

func (s *service) RecalculateCart(cartID uuid.UUID) error {
	query, args, err := QB.
		Select("cart_items.quantity, items.price").
		From("cart_items").
		Join("items ON cart_items.item_id = items.id").
		Where("cart_items.cart_id = ?", cartID).
		ToSql()
	if err != nil {
		return err
	}

	var cartItems []struct {
		Quantity int     `db:"quantity"`
		Price    float64 `db:"price"`
	}

	if err := s.db.Select(&cartItems, query, args...); err != nil {
		return err
	}

	var totalPrice float64
	var totalQuantity int
	for _, item := range cartItems {
		totalPrice += float64(item.Quantity) * item.Price
		totalQuantity += item.Quantity
	}

	query, args, err = QB.Update("carts").
		Set("total_price", totalPrice).
		Set("quantity", totalQuantity).
		Set("updated_at", time.Now()).
		Where("id = ?", cartID).
		ToSql()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(query, args...)
	return err
}

func (s *service) EmptyCart(userID string) error {
	cart, err := s.GetCart(userID)
	if err != nil {
		return err
	}

	if err := s.ClearCartItems(cart.ID); err != nil {
		return err
	}

	query, args, err := QB.Update("carts").
		Set("total_price", 0).
		Set("quantity", 0).
		Set("vendor_id", nil).
		Set("updated_at", time.Now()).
		Where("id = ?", cart.ID).
		ToSql()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(query, args...)
	return err
}

func (s *service) ProcessCheckout(cart types.Cart) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	order := types.Order{
		ID:             uuid.New(),
		TotalOrderCost: cart.TotalPrice,
		VendorId:       cart.VendorId,
		CustomerId:     cart.ID,
		Status:         "preparing",
		Created_at:     time.Now(),
		Updated_at:     time.Now(),
	}

	if err := s.CreateOrder(tx, order); err != nil {
		return err
	}

	if err := s.CreateOrderItems(tx, order.ID, cart.ID); err != nil {
		return err
	}

	if err := s.ClearCartItems(cart.ID); err != nil {
		return err
	}

	if err := s.ResetCartAfterCheckout(cart.ID); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *service) CreateOrder(tx *sqlx.Tx, order types.Order) error {
	query, args, err := QB.Insert("orders").
		Columns("id", "total_order_cost", "vendor_id", "customer_id", "status", "created_at", "updated_at").
		Values(order.ID, order.TotalOrderCost, order.VendorId, order.CustomerId, order.Status, order.Created_at, order.Updated_at).
		ToSql()
	if err != nil {
		return err
	}
	_, err = tx.Exec(query, args...)
	return err
}

func (s *service) CreateOrderItems(tx *sqlx.Tx, orderID, cartID uuid.UUID) error {
	cartItems, err := s.GetCartItems(cartID)
	if err != nil {
		return err
	}

	for _, item := range cartItems {
		itemDetails, err := s.GetCartItem(item.ItemId)
		if err != nil {
			return err
		}

		query, args, err := QB.Insert("order_items").
			Columns("order_id", "item_id", "quantity", "price").
			Values(orderID, item.ItemId, item.Quantity, itemDetails.Price).
			ToSql()
		if err != nil {
			return err
		}

		_, err = tx.Exec(query, args...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *service) ResetCartAfterCheckout(cartID uuid.UUID) error {
	query, args, err := QB.Update("carts").
		Set("total_price", 0).
		Set("quantity", 0).
		Set("vendor_id", nil).
		Set("updated_at", time.Now()).
		Where("id = ?", cartID).
		ToSql()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(query, args...)
	return err
}

type contextKey string

var UserIDKey = contextKey("userID")

func (s *service) GetUserID(r *http.Request) string {
	return r.Context().Value(UserIDKey).(string)
}

func (s *service) ParseAddCartParams(r *http.Request) (uuid.UUID, int, error) {
	itemID, err := uuid.Parse(r.FormValue("item_id"))
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("invalid item ID")
	}

	quantity, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil || quantity <= 0 {
		return uuid.Nil, 0, fmt.Errorf("quantity must be a positive integer")
	}

	return itemID, quantity, nil
}
