package domain

import (
	"context"
	"log"
	"route256/loms/internal/domain/models"
	sender "route256/loms/internal/sender"
)

func (m *Model) CreateOrder(ctx context.Context, order models.Order) (int64, error) {

	orderId, err := m.repo.CreateOrder(ctx, order)
	if err != nil {
		log.Printf("CreateOrder err %v\n", err.Error())
		return 0, err
	}

	err = m.sender.SendMessage(sender.PaymentMessage{
		User:    order.UserId,
		OrderID: orderId,
		Status:  OrderNew,
	})
	if err != nil {
		log.Printf("CreateOrder err %v\n", err.Error())
		return 0, err
	}

	return int64(orderId), nil
}
