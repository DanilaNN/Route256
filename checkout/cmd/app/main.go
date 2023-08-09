package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"route256/checkout/internal/api/checkout"
	loms_client "route256/checkout/internal/clients/loms"
	product_service_grpc "route256/checkout/internal/clients/product_service"
	"route256/checkout/internal/config"
	"route256/checkout/internal/domain"
	"route256/checkout/internal/pkg/logger"
	"route256/checkout/internal/pkg/metrics"
	"route256/checkout/internal/pkg/ratelimit"
	"route256/checkout/internal/pkg/tracer"
	postgres "route256/checkout/internal/repository"
	"route256/checkout/internal/repository/postgress/tx"
	"route256/checkout/pkg/checkout_v1"
	closer "route256/libs/pgcloser"
	"syscall"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/aitsvet/debugcharts"
	_ "github.com/aitsvet/debugcharts"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

const targetPSRPS = 10

func main() {
	sq.StatementBuilder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	var closer = new(closer.Closer)

	err := config.Init()
	if err != nil {
		log.Fatalln("ERR: ", err)
	}

	// Init logger
	logger.SetLoggerByEnvironment(config.AppConfig.Env)
	logger.Info("Start Checkout")

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

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.AppConfig.PortGrpc))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	lomsClient, err := loms_client.NewClient(config.AppConfig.Services.LomsGrpc)
	if err != nil {
		return fmt.Errorf("failed to create loms Client: %v", err)
	}

	prodServClient, err := product_service_grpc.NewClient(
		config.AppConfig.Services.ProductServGrpc,
		config.AppConfig.Token,
		ratelimit.New(targetPSRPS))
	if err != nil {
		return fmt.Errorf("failed to create product service Client: %v", err)
	}

	pool, err := pgxpool.Connect(ctx, config.AppConfig.PgConnStr)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	closer.Add(func(ctx context.Context) error {
		pool.Close()
		return nil
	})

	provider := tx.New(pool)
	repo := postgres.New(provider)

	model := domain.New(lomsClient, prodServClient, repo, provider)

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logger.MiddlewareGRPC,
			tracer.MiddlewareGRPC,
			metrics.MiddlewareGRPC,
		),
	)
	reflection.Register(s)
	checkout_v1.RegisterCheckoutServer(s, checkout.NewCheckoutServer(model))

	log.Printf("server listening at %v", lis.Addr())

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
		log.Fatalln("Failed to dial server:", err)
	}

	mux := runtime.NewServeMux()
	err = checkout_v1.RegisterCheckoutHandler(context.Background(), mux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	gwServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.AppConfig.Port),
		Handler: mux,
	}

	// debug
	go func() {
		t := time.NewTicker(time.Second)
		for range t.C {
			debugcharts.RPS.Set(0)
		}
	}()

	go func() {
		log.Println(http.ListenAndServe("checkout:6060", nil))
	}()

	log.Printf("Serving gRPC-Gateway on %s\n", gwServer.Addr)
	err = gwServer.ListenAndServe()
	if err != nil {
		return fmt.Errorf("failed gwServer.ListenAndServe(): %v", err)
	}

	return nil
}
