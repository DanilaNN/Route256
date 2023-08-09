package domain

import (
	"context"
	"route256/loms/internal/domain/models"
)

func (m *Model) Stocks(ctx context.Context, sku uint32) ([]models.Stock, error) {

	// mock
	stocks := make([]models.Stock, 0, 1)
	stocks = append(stocks, models.Stock{WarehouseID: 10, Count: 20})

	return stocks, nil
}
