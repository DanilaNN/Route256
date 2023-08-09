package domain

import (
	"context"
	"fmt"
	"route256/checkout/internal/domain/models"
)

func (m *Model) DeleteFromCart(ctx context.Context, user int64, sku uint32, count uint16) error {

	userOrder := models.UserOrderItem{
		User: user,
		Item: models.OrderItem{
			Sku:   uint64(sku),
			Count: count,
		},
	}

	// Transaction on business level
	err := m.tx.RunRepeatableRead(ctx, func(ctxTx context.Context) error {
		count, err := m.repo.GetSkuCountInCart(ctxTx, userOrder)
		if err != nil {
			return err
		}

		if count > uint32(userOrder.Item.Count) {
			delta := int32(count - uint32(userOrder.Item.Count))
			err = m.repo.DecreaseSkuCountInCart(ctxTx, userOrder, delta)
		} else {
			err = m.repo.DeleteUserSku(ctxTx, userOrder)
		}

		return err
	})

	if err != nil {
		return fmt.Errorf("DeleteFromCart: %w", err)
	}

	return nil
}
