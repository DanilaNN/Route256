package server

import (
	"route256/loms/internal/domain/models"
	"route256/loms/pkg/loms_v1"
)

func CreateOrderInfoFromReq(req *loms_v1.CreateOrderRequest) (models.Order, error) {
	err := req.ValidateAll()
	if err != nil {
		return models.Order{}, err
	}

	order := models.Order{UserId: req.GetUser()}
	order.Items = make([]models.OrderItem, 0, len(req.GetItems()))

	orderItemsReq := req.GetItems()
	for _, orderItemReq := range orderItemsReq {
		order.Items = append(order.Items, models.OrderItem{Sku: uint64(orderItemReq.Sku), Count: uint16(orderItemReq.Count)})
	}

	return order, nil
}

func OrderToResponse(item models.OrderItem) *loms_v1.OrderItem {
	return &loms_v1.OrderItem{
		Sku:   uint32(item.Sku),
		Count: uint32(item.Count),
	}
}

func ListOrderToResponse(order models.Order) *loms_v1.ListOrderResponse {
	resp := loms_v1.ListOrderResponse{Status: order.Status, User: order.UserId}
	for _, orderItem := range order.Items {
		resp.Items = append(resp.Items, OrderToResponse(orderItem))
	}
	return &resp
}

func StockToResponse(stock models.Stock) *loms_v1.Stock {
	return &loms_v1.Stock{
		WarehouseId: stock.WarehouseID,
		Count:       stock.Count,
	}
}

func StocksToResponse(stocks []models.Stock) *loms_v1.StocksResponse {
	resp := loms_v1.StocksResponse{}
	for _, stock := range stocks {
		resp.Stocks = append(resp.Stocks, StockToResponse(stock))
	}
	return &resp
}
