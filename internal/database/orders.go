package database

import (
	"net/url"
	"restaurant-management-backend/internal/types"
)

func (s *service) FetchOrders(queryParams map[string][]string) ([]types.Order, types.Meta, error) {
	var orders []types.Order

	urlValues := make(url.Values)
	for key, values := range queryParams {
		for _, value := range values {
			urlValues.Add(key, value)
		}
	}

	columns := []string{
		"id", "total_order_cost", "vendor_id", "customer_id", "status", "created_at", "updated_at",
	}

	searchColumns := []string{"id", "status"}

	meta, err := s.BuildQuery(
		&orders,
		"orders",
		[]string{},
		columns,
		searchColumns,
		urlValues,
		[]string{},
	)

	if err != nil {
		return nil, types.Meta{}, err
	}

	if orders == nil {
		orders = []types.Order{}
	}

	return orders, *meta, nil
}

func (s *service) EnrichOrdersWithItems(orders []types.Order) error {
	for i := range orders {
		err := s.AttachOrderItems(&orders[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *service) FetchOrder(id string) (types.Order, error) {
	var order types.Order
	query, args, err := QB.Select("*").From("orders").Where("id = ?", id).ToSql()
	if err != nil {
		return order, err
	}
	err = s.db.Get(&order, query, args...)
	return order, err
}

func (s *service) AttachOrderItems(order *types.Order) error {
	var orderItems []types.OrderItems
	query, args, err := QB.Select("*").From("order_items").Where("order_id = ?", order.ID).ToSql()
	if err != nil {
		return err
	}
	err = s.db.Select(&orderItems, query, args...)
	if err != nil {
		return err
	}
	order.OrderItems = orderItems
	return nil
}

func (s *service) UpdateOrderStatus(id, status string) error {
	query, args, err := QB.Update("orders").Set("status", status).Where("id = ?", id).ToSql()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(query, args...)
	return err
}
