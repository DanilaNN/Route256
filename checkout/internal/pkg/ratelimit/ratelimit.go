package ratelimit

import (
	"context"

	"golang.org/x/time/rate"
)

type Limiter struct {
	l *rate.Limiter
}

func (lim *Limiter) Wait(ctx context.Context) error {
	return lim.l.Wait(ctx)
}

func New(rps int) *Limiter {
	return &Limiter{l: rate.NewLimiter(rate.Limit(rps), 1)}
}
