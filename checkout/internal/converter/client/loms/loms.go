package loms_converter

import (
	"route256/checkout/internal/domain/models"
	"route256/checkout/pkg/loms_v1"
)

func OrderToReq(order models.Order) *loms_v1.CreateOrderRequest {

	reqOrder := loms_v1.CreateOrderRequest{User: order.UserId}
	reqOrder.Items = make([]*loms_v1.OrderItem, 0, len(order.Items))

	for _, item := range order.Items {
		reqOrder.Items = append(reqOrder.Items, &loms_v1.OrderItem{Sku: uint32(item.Sku),
			Count: uint32(item.Count)})
	}

	return &reqOrder
}
