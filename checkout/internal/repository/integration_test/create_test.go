//go:build integration

package integration_test

import (
	"context"
	"route256/checkout/internal/domain/models"
)

const (
	userId = int64(1)
	sku    = uint64(1121)
	count  = 5
)

func (s *Suite) TestAddCart() {
	var cartsTarget models.UserOrderItem
	cartsTarget.User = userId
	cartsTarget.Item.Count = count
	cartsTarget.Item.Sku = sku
	ctx := context.Background()

	// Act
	err := s.repo.AddCart(ctx, cartsTarget)

	// Assert
	s.Require().NoError(err)
	s.Require().Condition(func() (success bool) {
		var user int64
		var sku uint64
		var count uint16
		err := s.pool.QueryRow(ctx, "SELECT user_id, sku, count FROM carts WHERE user_id=$1", cartsTarget.User).Scan(&user, &sku, &count)
		s.Require().NoError(err)

		s.Require().Equal(user, cartsTarget.User)
		s.Require().Equal(count, cartsTarget.Item.Count)
		s.Require().Equal(sku, cartsTarget.Item.Sku)

		return true
	})
}

func (s *Suite) TestGetSkuCountInCart() {
	var cartsTarget models.UserOrderItem
	cartsTarget.User = userId
	cartsTarget.Item.Count = count
	cartsTarget.Item.Sku = sku
	ctx := context.Background()

	_, err := s.pool.Exec(ctx, "INSERT INTO carts (user_id,sku,count) VALUES ($1,$2,$3)", userId, sku, count)
	s.Require().NoError(err)

	// Act
	countActual, err := s.repo.GetSkuCountInCart(ctx, cartsTarget)

	// Assert
	s.Require().NoError(err)
	s.Require().Equal(countActual, uint32(count))
}

func (s *Suite) TestDecreaseSkuCountInCart() {

	var cartsTarget models.UserOrderItem
	cartsTarget.User = userId
	cartsTarget.Item.Count = count
	cartsTarget.Item.Sku = sku
	ctx := context.Background()
	const delta = count

	_, err := s.pool.Exec(ctx, "INSERT INTO carts (user_id,sku,count) VALUES ($1,$2,$3)", userId, sku, count*2)
	s.Require().NoError(err)

	// Act
	err = s.repo.DecreaseSkuCountInCart(ctx, cartsTarget, int32(delta))

	// Assert
	s.Require().NoError(err)
	s.Require().Condition(func() (success bool) {
		var countActual int64
		err := s.pool.QueryRow(ctx, "SELECT count FROM carts WHERE user_id=$1", userId).Scan(&countActual)
		s.Require().NoError(err)
		s.Require().Equal(countActual, int64(count))
		return true
	})
}

func (s *Suite) TestDeleteUserSku() {

	var cartsTarget models.UserOrderItem
	cartsTarget.User = userId
	cartsTarget.Item.Count = count
	cartsTarget.Item.Sku = sku
	ctx := context.Background()

	_, err := s.pool.Exec(ctx, "INSERT INTO carts (user_id,sku,count) VALUES ($1,$2,$3)", userId, sku, count*2)
	s.Require().NoError(err)

	// Act
	err = s.repo.DeleteUserSku(ctx, cartsTarget)

	// Assert
	s.Require().NoError(err)
	s.Require().Condition(func() (success bool) {
		var countActual int64
		err := s.pool.QueryRow(ctx, "SELECT count FROM carts WHERE user_id=$1", userId).Scan(&countActual)
		s.Require().ErrorContains(err, "no rows in result set")
		return true
	})

}

func (s *Suite) TestGetCart() {

	const sku1 = 5
	const sku2 = 6
	const count1 = 10
	const count2 = 11
	ctx := context.Background()

	var testCart models.ProductCarts
	testCart.Carts = append(testCart.Carts, models.ProductCart{
		Sku:   sku1,
		Count: count1,
	})
	testCart.Carts = append(testCart.Carts, models.ProductCart{
		Sku:   sku2,
		Count: count2,
	})

	_, err := s.pool.Exec(ctx, "INSERT INTO carts (user_id,sku,count) VALUES ($1,$2,$3), ($4,$5,$6)", userId, sku1, count1, userId, sku2, count2)
	s.Require().NoError(err)

	// Act
	carts, err := s.repo.GetCart(context.Background(), userId)

	// Assert
	s.Require().NoError(err)
	s.Require().ElementsMatch(carts.Carts, testCart.Carts)

}

func (s *Suite) TestDeleteCart() {

	const sku1 = 5
	const sku2 = 6
	const count1 = 10
	const count2 = 11
	ctx := context.Background()

	_, err := s.pool.Exec(ctx, "INSERT INTO carts (user_id,sku,count) VALUES ($1,$2,$3), ($4,$5,$6)", userId, sku1, count1, userId, sku2, count2)
	s.Require().NoError(err)

	// Act
	err = s.repo.DeleteCart(context.Background(), userId)

	// Assert
	s.Require().NoError(err)
	s.Require().Condition(func() (success bool) {
		var countActual int64
		err := s.pool.QueryRow(ctx, "SELECT count FROM carts WHERE user_id=$1", userId).Scan(&countActual)
		s.Require().ErrorContains(err, "no rows in result set")
		return true
	})
}
