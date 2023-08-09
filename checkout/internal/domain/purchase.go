package domain

import (
	"context"
	"route256/checkout/internal/domain/models"
)

func (m *Model) Purchase(ctx context.Context, userId int64) (int64, error) {

	productCarts, err := m.repo.GetCart(ctx, userId)
	if err != nil {
		return 0, nil
	}

	var order models.Order
	order.UserId = userId

	for _, item := range productCarts.Carts {
		order.Items = append(order.Items, models.OrderItem{Sku: item.Sku, Count: item.Count})
	}

	orderId, err := m.lomsClient.CreateOrder(ctx, order)
	if err != nil {
		return 0, err
	}

	err = m.repo.DeleteCart(ctx, userId)
	if err != nil {
		return 0, nil
	}

	return orderId, nil
}
