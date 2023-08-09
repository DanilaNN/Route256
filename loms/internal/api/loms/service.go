package loms

import (
	"context"
	"route256/loms/internal/converter/server"
	"route256/loms/internal/domain"
	"route256/loms/pkg/loms_v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ loms_v1.LomsServer = (*Service)(nil)

type Service struct {
	loms_v1.UnimplementedLomsServer
	Model *domain.Model
}

func (s *Service) CreateOrder(ctx context.Context, req *loms_v1.CreateOrderRequest) (*loms_v1.CreateOrderResponse, error) {

	order, err := server.CreateOrderInfoFromReq(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	orderID, err := s.Model.CreateOrder(ctx, order)
	if err != nil {
		return nil, err
	}
	return &loms_v1.CreateOrderResponse{OrderId: orderID}, err
}

func (s *Service) ListOrder(ctx context.Context, req *loms_v1.ListOrderRequest) (*loms_v1.ListOrderResponse, error) {

	err := req.ValidateAll()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	orderInfo, err := s.Model.ListOrder(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}

	return server.ListOrderToResponse(orderInfo), nil
}

func (s *Service) OrderPayed(ctx context.Context, req *loms_v1.OrderPayedRequest) (*emptypb.Empty, error) {

	err := req.ValidateAll()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	err = s.Model.OrderPayed(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, err
}

func (s *Service) CancelOrder(ctx context.Context, req *loms_v1.CancelOrderRequest) (*emptypb.Empty, error) {

	err := req.ValidateAll()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	err = s.Model.CancelOrder(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, err
}

func (s *Service) Stocks(ctx context.Context, req *loms_v1.StocksRequest) (*loms_v1.StocksResponse, error) {

	err := req.ValidateAll()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	stocks, err := s.Model.Stocks(ctx, req.Sku)
	if err != nil {
		return nil, err
	}
	return server.StocksToResponse(stocks), nil
}

func NewLomsServer(model *domain.Model) *Service {
	return &Service{Model: model}
}
