package main

import (
	"context"
	"log"
	"os/signal"
	"route256/notifications/cmd/app"
	"syscall"
)

func main() {
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGUSR1, syscall.SIGTERM)
	defer stop()

	a, err := app.New(ctx)
	if err != nil {
		log.Fatalf("cannot create app: %s", err.Error())
	}

	if err = a.Run(ctx); err != nil {
		log.Fatalf("cannot run app: %s", err.Error())
	}
}
