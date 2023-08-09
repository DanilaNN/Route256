package domain

import (
	"context"
	"errors"
	"route256/checkout/internal/domain/mocks"
	"route256/checkout/internal/domain/models"
	"testing"

	"github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/require"
)

func Test_ListCart(t *testing.T) {
	t.Parallel()

	const (
		sku1   = uint32(1122)
		sku2   = uint32(1123)
		price1 = 2000
		price2 = 3000
		count1 = 5
		count2 = 10

		userId = int64(1)
	)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		pgrepository := mocks.NewPGRepository(t)
		psClient := mocks.NewProductServiceClient(t)

		brand1 := gofakeit.Name()
		brand2 := gofakeit.Name()
		cart1 := models.ProductCart{Sku: uint64(sku1), Count: count1}
		cart2 := models.ProductCart{Sku: uint64(sku2), Count: count2}
		cartsStub := make([]models.ProductCart, 0, 2)
		cartsStub = append(cartsStub, cart1)
		cartsStub = append(cartsStub, cart2)

		ctx := context.Background()

		productCarts := models.ProductCarts{Carts: cartsStub, TotalPrice: 0}

		pgrepository.On("GetCart", ctx, userId).Return(productCarts, nil).Once()

		psClient.On("GetProduct", ctx, sku1).Return(models.Product{Name: brand1, Price: price1}, nil).Once()
		psClient.On("GetProduct", ctx, sku2).Return(models.Product{Name: brand2, Price: price2}, nil).Once()

		// prepare target answer
		cartsStub2 := make([]models.ProductCart, 0, 2)
		cartsStub2 = append(cartsStub2, cart1)
		cartsStub2 = append(cartsStub2, cart2)
		cartsStub2[0].Name = brand1
		cartsStub2[1].Name = brand2
		cartsStub2[0].Price = price1
		cartsStub2[1].Price = price2
		targetCarts := models.ProductCarts{Carts: cartsStub2, TotalPrice: 0}

		// Act
		carts, err := (&Model{
			productServiceClient: psClient,
			repo:                 pgrepository,
		}).ListCart(ctx, userId)

		// Assert
		require.NoError(t, err)
		require.Equal(t, carts.TotalPrice, uint32(price1+price2), "totalPrice")
		require.Len(t, carts.Carts, 2, "len carts != 2")
		require.ElementsMatch(t, carts.Carts, targetCarts.Carts)
	})

	t.Run("error while getting carts from storage", func(t *testing.T) {
		t.Parallel()

		errStub := errors.New("stub")

		const (
			userId = int64(1)
		)
		ctx := context.Background()

		pgrepository := mocks.NewPGRepository(t)
		pgrepository.On("GetCart", ctx, userId).Return(models.ProductCarts{}, errStub).Once()

		// Act
		_, err := (&Model{
			repo: pgrepository,
		}).ListCart(ctx, userId)

		// Assert
		require.ErrorIs(t, err, errStub)
	})

	t.Run("error requesting Product Service", func(t *testing.T) {
		t.Parallel()

		errStub := errors.New("stub")

		pgrepository := mocks.NewPGRepository(t)
		psClient := mocks.NewProductServiceClient(t)

		cart1 := models.ProductCart{Sku: uint64(sku1), Count: count1}
		cart2 := models.ProductCart{Sku: uint64(sku2), Count: count2}
		cartsStub := make([]models.ProductCart, 0, 2)
		cartsStub = append(cartsStub, cart1)
		cartsStub = append(cartsStub, cart2)

		ctx := context.Background()

		productCarts := models.ProductCarts{Carts: cartsStub, TotalPrice: 0}

		pgrepository.On("GetCart", ctx, userId).Return(productCarts, nil).Once()

		psClient.On("GetProduct", ctx, sku1).Return(models.Product{}, errStub).Once()
		psClient.On("GetProduct", ctx, sku2).Return(models.Product{}, errStub).Once()

		// prepare target answer
		targetCarts := models.ProductCarts{Carts: cartsStub, TotalPrice: 0}

		// Act
		carts, err := (&Model{
			repo:                 pgrepository,
			productServiceClient: psClient,
		}).ListCart(ctx, userId)

		// Assert
		require.NoError(t, err)
		require.ElementsMatch(t, targetCarts.Carts, carts.Carts)
	})
}
