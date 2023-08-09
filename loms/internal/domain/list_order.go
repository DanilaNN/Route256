package domain

import (
	"context"
	"route256/loms/internal/domain/models"
)

func (m *Model) ListOrder(ctx context.Context, orderId int64) (models.Order, error) {
	return m.repo.GetOrder(ctx, orderId)
}
