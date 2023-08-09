package domain

import (
	"context"
	sender "route256/loms/internal/sender"
)

func (m *Model) CancelOrder(ctx context.Context, orderId int64) error {

	err := m.repo.SetOrderStatus(ctx, orderId, OrderCancelled)
	if err != nil {
		return err
	}

	order, err := m.repo.GetOrder(ctx, orderId)
	if err != nil {
		return err
	}

	return m.sender.SendMessage(sender.PaymentMessage{
		User:    order.UserId,
		OrderID: uint64(orderId),
		Status:  OrderCancelled,
	})
}
