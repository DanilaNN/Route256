package checkout

import (
	"context"
	"route256/checkout/internal/converter/server"
	"route256/checkout/internal/domain"
	"route256/checkout/pkg/checkout_v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ checkout_v1.CheckoutServer = (*Service)(nil)

type Service struct {
	checkout_v1.UnimplementedCheckoutServer
	Model *domain.Model
	//impl *checkout.Service
}

func (s *Service) AddToCart(ctx context.Context, req *checkout_v1.AddToCartRequest) (*emptypb.Empty, error) {
	err := req.ValidateAll()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	user, sku, count, err := server.AddToCartInfoFromReq(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	err = s.Model.AddToCart(ctx, user, sku, uint16(count))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) DeleteFromCart(ctx context.Context, req *checkout_v1.DeleteFromCartRequest) (*emptypb.Empty, error) {
	err := req.ValidateAll()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	user, sku, count, err := server.DeleteFromCartInfoFromReq(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	err = s.Model.DeleteFromCart(ctx, user, sku, uint16(count))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) ListCart(ctx context.Context, req *checkout_v1.ListCartRequest) (*checkout_v1.ListCartResponse, error) {
	err := req.ValidateAll()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	products, err := s.Model.ListCart(ctx, req.User)
	if err != nil {
		return nil, err
	}

	response := server.ListCartToResponse(products)
	return response, nil
}

func (s *Service) Purchase(ctx context.Context, req *checkout_v1.PurchaseRequest) (*checkout_v1.PurchaseResponse, error) {
	err := req.ValidateAll()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	orderId, err := s.Model.Purchase(ctx, req.User)
	if err != nil {
		return nil, err
	}
	return &checkout_v1.PurchaseResponse{OrderId: orderId}, nil
}

func NewCheckoutServer(model *domain.Model) *Service {
	return &Service{Model: model}
}
