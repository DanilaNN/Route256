package loms_client

import (
	"context"
	"log"
	"route256/checkout/internal/domain/models"
	"route256/checkout/pkg/loms_v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	loms_v1.LomsClient
	conn  *grpc.ClientConn
	Token string
}

func (s *Client) Stocks(ctx context.Context, sku uint32) ([]models.Stock, error) {

	resp, err := s.LomsClient.Stocks(ctx, &loms_v1.StocksRequest{Sku: sku})
	if err != nil {
		return []models.Stock{}, err
	}

	result := make([]models.Stock, 0, len(resp.GetStocks()))
	for _, v := range resp.GetStocks() {
		result = append(result, models.Stock{
			WarehouseID: v.WarehouseId,
			Count:       v.Count,
		})
	}

	return result, nil
}

func (s *Client) CreateOrder(ctx context.Context, order models.Order) (int64, error) {

	// TODO: use loms_converter
	reqOrder := loms_v1.CreateOrderRequest{User: order.UserId}
	reqOrder.Items = make([]*loms_v1.OrderItem, 0, len(order.Items))

	for _, item := range order.Items {
		reqOrder.Items = append(reqOrder.Items, &loms_v1.OrderItem{Sku: uint32(item.Sku),
			Count: uint32(item.Count)})
	}

	resp, err := s.LomsClient.CreateOrder(ctx, &reqOrder)
	if err != nil {
		return 0, err
	}

	return resp.OrderId, nil
}

func NewClient(target string) (*Client, error) {
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to connect to loms server: %v", err)
		return nil, err
	}

	c := loms_v1.NewLomsClient(conn)
	return &Client{LomsClient: c, conn: conn}, nil
}
