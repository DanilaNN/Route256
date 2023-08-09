package domain

import (
	"context"
	"route256/loms/internal/domain/models"
	sender "route256/loms/internal/sender"
)

const ServiceName = "loms"

const (
	OrderNew       = "new"
	OrderPayed     = "payed"
	OrderCancelled = "cancelled"
)

type Model struct {
	repo   Repository
	sender Sender
}

type Repository interface {
	CreateOrder(ctx context.Context, order models.Order) (uint64, error)
	GetOrder(ctx context.Context, orderID int64) (models.Order, error)
	SetOrderStatus(ctx context.Context, orderID int64, status string) error
}

type Sender interface {
	SendMessage(message sender.PaymentMessage) error
	SendMessages(messages []sender.PaymentMessage) error
}

func New(repo Repository, sender Sender) *Model {
	return &Model{
		repo:   repo,
		sender: sender,
	}
}
