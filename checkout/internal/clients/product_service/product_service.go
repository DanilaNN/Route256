package product_service_grpc

import (
	"context"
	"log"
	"route256/checkout/internal/domain"
	"route256/checkout/internal/domain/models"
	productService_v1 "route256/checkout/pkg/productservice_v1"

	"github.com/aitsvet/debugcharts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	productService_v1.ProductServiceClient
	conn    *grpc.ClientConn
	limiter domain.Limiter
	Token   string
}

func (s *Client) GetProduct(ctx context.Context, sku uint32) (models.Product, error) {
	debugcharts.RPS.Add(1)
	err := s.limiter.Wait(ctx) // Honors the rate limit
	if err != nil {
		return models.Product{}, err
	}
	resp, err := s.ProductServiceClient.GetProduct(ctx, &productService_v1.GetProductRequest{Token: s.Token, Sku: sku})
	if err != nil {
		return models.Product{}, err
	}

	return models.Product{Name: resp.GetName(), Price: int32(resp.GetPrice())}, nil
}

func (s *Client) ListSkus(ctx context.Context, startAfterSku uint32, count uint32) (models.Skus, error) {

	resp, err := s.ProductServiceClient.ListSkus(ctx, &productService_v1.ListSkusRequest{Token: s.Token,
		StartAfterSku: startAfterSku, Count: count})
	if err != nil {
		return models.Skus{}, err
	}

	skus := make(models.Skus, 0, len(resp.Skus))
	for _, sku := range resp.Skus {
		skus = append(skus, sku)
	}
	return skus, nil
}

func NewClient(target string, token string, l domain.Limiter) (*Client, error) {
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to connect to product-service server: %v", err)
		return nil, err
	}

	c := productService_v1.NewProductServiceClient(conn)
	return &Client{
		ProductServiceClient: c,
		Token:                token,
		conn:                 conn,
		limiter:              l,
	}, nil
}

// TODO: Maybe use this constructor?
// func NewClient(c productService_v1.ProductServiceClient, token string) *Client {
// 	return &Client{ProductServiceClient: c, Token: token}
// }
