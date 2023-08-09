package server

import (
	"route256/checkout/internal/domain/models"
	"route256/checkout/pkg/checkout_v1"
)

func AddToCartInfoFromReq(req *checkout_v1.AddToCartRequest) (int64, uint32, uint32, error) {
	var user int64
	var sku uint32
	var count uint32

	user = req.GetUser()
	sku = req.GetSku()
	count = req.GetCount()

	return user, sku, count, nil
}

func DeleteFromCartInfoFromReq(req *checkout_v1.DeleteFromCartRequest) (int64, uint32, uint32, error) {
	var user int64
	var sku uint32
	var count uint32

	user = req.GetUser()
	sku = req.GetSku()
	count = req.GetCount()

	return user, sku, count, nil
}

func CartToResponse(product models.ProductCart) *checkout_v1.ListCartItem {
	return &checkout_v1.ListCartItem{
		Sku:   uint32(product.Sku),
		Count: uint32(product.Count),
		Name:  product.Name,
		Price: product.Price,
	}
}

func ListCartToResponse(products models.ProductCarts) *checkout_v1.ListCartResponse {

	resp := &checkout_v1.ListCartResponse{}
	for _, product := range products.Carts {
		resp.Items = append(resp.Items, CartToResponse(product))
	}
	resp.TotalPrice = products.TotalPrice

	return resp

}
