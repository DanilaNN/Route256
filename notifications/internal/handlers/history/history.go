package history

import (
	"context"
	"errors"
	"route256/notifications/internal/domain"
	"time"
)

type Handler struct {
	Model *domain.Model
}

type Response struct {
	Statuses []string `json:"statuses"`
}

type Request struct {
	User      int64     `json:"user"`
	TimeStart time.Time `json:"timeStart"`
	TimeStop  time.Time `json:"timeStop"`
}

var (
	ErrInvalidUserId = errors.New("invalid user id")
	CacheMis         = errors.New("getting value from cache")
)

func (r Request) Validate() error {
	if r.User == 0 {
		return ErrInvalidUserId
	}
	return nil
}

func (h *Handler) Handle(ctx context.Context, req Request) (Response, error) {
	responses, err := h.Model.GetFromCache(req.User, req.TimeStart, req.TimeStop)
	if err != nil {
		return Response{}, CacheMis
	}

	return Response{responses}, nil
}
