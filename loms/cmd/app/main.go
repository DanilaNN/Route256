package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	closer "route256/libs/pgcloser"
	"route256/loms/internal/api/loms"
	"route256/loms/internal/config"
	"route256/loms/internal/domain"
	"route256/loms/internal/infrastructure/kafka"
	"route256/loms/internal/pkg/logger"
	"route256/loms/internal/pkg/metrics"
	"route256/loms/internal/pkg/tracer"
	"route256/loms/internal/repository"
	sender "route256/loms/internal/sender"
	"route256/loms/pkg/loms_v1"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	err := config.Init()
	if err != nil {
		return err
	}

	// Init logger
	logger.SetLoggerByEnvironment(config.AppConfig.Env)
	logger.Info("Start Loms")

	// Init tracer
	if err := tracer.InitGlobal(domain.ServiceName, config.AppConfig.JaegerHost); err != nil {
		return err
	}

	// Metrics
	go func() {
		mux1 := runtime.NewServeMux(
			runtime.WithOutgoingHeaderMatcher(func(key string) (string, bool) {
				switch key {
				case "x-trace-id":
					return key, true
				}
				return runtime.DefaultHeaderMatcher(key)
			}),
		)

		if err := mux1.HandlePath(http.MethodGet, "/metrics", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
			promhttp.Handler().ServeHTTP(w, r)
		}); err != nil {
			logger.Fatal("something wrong with metrics handler", err)
		}

		logger.Info("HTTP metrics server started on: ", config.AppConfig.MetricsHost)
		_ = http.ListenAndServe(config.AppConfig.MetricsHost, mux1)

	}()

	// PG repo
	pool, err := pgxpool.Connect(ctx, config.AppConfig.PgConnStr)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	var closer = new(closer.Closer)
	closer.Add(func(ctx context.Context) error {
		pool.Close()
		return nil
	})
	repo := repository.New(pool)

	// Kafka
	kafkaProducer, err := kafka.NewProducer(config.AppConfig.Brokers)
	if err != nil {
		fmt.Println(err)
	}
	sender := sender.New(kafkaProducer, "orders")

	// Service model
	model := domain.New(repo, sender)

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logger.MiddlewareGRPC,
			tracer.MiddlewareGRPC,
			metrics.MiddlewareGRPC,
		),
	)
	reflection.Register(s)
	loms_v1.RegisterLomsServer(s, loms.NewLomsServer(model))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.AppConfig.GrpcPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	logger.Info("server listening at %v", lis.Addr())

	go func() {
		if err = s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	conn, err := grpc.DialContext(
		context.Background(),
		lis.Addr().String(),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to dial server: %v", err)
	}

	mux := runtime.NewServeMux()
	err = loms_v1.RegisterLomsHandler(context.Background(), mux, conn)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %v", err)
	}

	gwServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.AppConfig.HttpPort),
		Handler: mux,
	}

	log.Printf("Serving gRPC-Gateway on %s\n", gwServer.Addr)
	err = gwServer.ListenAndServe()
	if err != nil {
		return fmt.Errorf("failed gwServer.ListenAndServe(): %v", err)
	}

	return nil
}
