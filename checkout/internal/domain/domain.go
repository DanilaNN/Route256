//go:generate mockery --filename product_service_mock.go --name ProductServiceClient
//go:generate mockery --filename pg_repository_mock.go --name PGRepository

package domain

import (
	"context"
	"route256/checkout/internal/domain/models"
)

const ServiceName = "Checkout"

type Model struct {
	lomsClient           LomsClient
	productServiceClient ProductServiceClient
	repo                 PGRepository
	tx                   TransactionManager
}

type LomsClient interface {
	Stocks(ctx context.Context, sku uint32) ([]models.Stock, error)
	CreateOrder(ctx context.Context, order models.Order) (int64, error)
}

type TransactionManager interface {
	RunRepeatableRead(ctx context.Context, fn func(ctxTx context.Context) error) error
}
type PGRepository interface {
	AddCart(ctx context.Context, userOrder models.UserOrderItem) error
	GetCart(ctx context.Context, orderID int64) (models.ProductCarts, error)
	DeleteCart(ctx context.Context, userID int64) error

	DecreaseSkuCountInCart(ctx context.Context, userOrder models.UserOrderItem, delta int32) error
	DeleteUserSku(ctx context.Context, userOrder models.UserOrderItem) error
	GetSkuCountInCart(ctx context.Context, userOrder models.UserOrderItem) (uint32, error)
}

type ProductServiceClient interface {
	GetProduct(ctx context.Context, sku uint32) (models.Product, error)
	ListSkus(ctx context.Context, startAfterSku uint32, count uint32) (models.Skus, error)
}

type Limiter interface {
	Wait(ctx context.Context) error
}

func New(LomsClient LomsClient, ProductServiceClient ProductServiceClient, Repo PGRepository, Tx TransactionManager) *Model {
	return &Model{
		lomsClient:           LomsClient,
		productServiceClient: ProductServiceClient,
		repo:                 Repo,
		tx:                   Tx,
	}
}
