package domain

import (
	"context"
	"log"
	"route256/checkout/internal/domain/models"
	"route256/checkout/internal/pkg/workerpool"
	"sync"

	"github.com/pkg/errors"
)

const poolSize = 5

func (m *Model) ListCart(ctx context.Context, userId int64) (models.ProductCarts, error) {

	productCarts, err := m.repo.GetCart(ctx, userId)
	if err != nil {
		return models.ProductCarts{}, errors.Wrap(err, "get cart from db")
	}

	wp := workerpool.NewFuture(ctx, poolSize,
		func(ctx context.Context, sku uint64) (*models.Product, error) {
			product, err := m.productServiceClient.GetProduct(ctx, uint32(sku))
			return &product, err
		})

	var mu sync.Mutex

	wg := sync.WaitGroup{}
	for idx, cart := range productCarts.Carts {
		wg.Add(1)
		p := wp.Exec(ctx, cart.Sku)
		i := idx
		go func() {
			result := <-p.Ch
			if result.Err != nil {
				log.Printf("Can not request PS to get price/name: %s", result.Err.Error())
				wg.Done()
				return
			}

			{
				mu.Lock()
				defer mu.Unlock()
				productCarts.TotalPrice += uint32(result.Value.Price)
				productCarts.Carts[i].Name = result.Value.Name
				productCarts.Carts[i].Price = uint32(result.Value.Price)
			}

			wg.Done()
		}()
	}
	wg.Wait()

	// TODO: maybe listSkus will be needed
	// skus, err := m.productServiceClient.ListSkus(ctx, 0, 10)

	return productCarts, nil
}
