package domain

import (
	"context"
	"errors"
	"fmt"
	"log"
	"route256/checkout/internal/domain/models"
)

var (
	ErrStockInsufficient = errors.New("stock insufficient")
)

func (m *Model) AddToCart(ctx context.Context, user int64, sku uint32, count uint16) error {

	stocks, err := m.lomsClient.Stocks(ctx, sku)
	if err != nil {
		return fmt.Errorf("get stocks: %w", err)
	}

	log.Printf("stocks: %+v", stocks)

	counter := uint64(0)
	for _, stock := range stocks {
		counter += stock.Count
		if counter > uint64(count) {
			break
		}
	}

	if counter > uint64(count) {
		err := m.repo.AddCart(ctx, models.UserOrderItem{
			User: user,
			Item: models.OrderItem{
				Sku:   uint64(sku),
				Count: count}})
		if err != nil {
			return fmt.Errorf("save product cart to db: %w", err)
		}
		return nil
	}

	return ErrStockInsufficient
}
