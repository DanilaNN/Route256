package app

import (
	"context"
	"log"
	"net/http"
	"route256/libs/srvwrapper"
	"route256/notifications/internal/cache"
	"route256/notifications/internal/config"
	"route256/notifications/internal/consumer"
	"route256/notifications/internal/domain"
	"route256/notifications/internal/handlers/history"
	"route256/notifications/internal/infrastructure/kafka"
	"route256/notifications/internal/pkg/logger"
	"route256/notifications/internal/pkg/tracer"
	"route256/notifications/internal/sender"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

type App struct {
	model *domain.Model
}

func New(ctx context.Context) (*App, error) {
	err := config.Init()
	if err != nil {
		return nil, err
	}

	bot, err := tgbotapi.NewBotAPI(config.AppConfig.TelegramToken)
	if err != nil {
		return nil, err
	}

	sender := sender.NewTelegramSender(bot, config.AppConfig.TelegramChatId)

	consumerGroup := kafka.NewConsumerGroup()
	consumer, err := consumer.New(&consumerGroup, config.AppConfig.ConsumerGroup, config.AppConfig.Brokers)
	if err != nil {
		return nil, err
	}

	cache := cache.NewLRU[domain.CacheKey](config.AppConfig.CacheCapacity)

	model := domain.New(sender, consumer, cache)

	// Init logger
	logger.SetLoggerByEnvironment(config.AppConfig.Env)
	logger.Info("Start Checkout")

	// Init tracer
	if err := tracer.InitGlobal(domain.ServiceName, config.AppConfig.JaegerHost); err != nil {
		return nil, err
	}

	// start server
	go func() {
		log.Printf("Starting HTTP server at port: %s", config.AppConfig.HttpPort)
		handHistory := &history.Handler{Model: model}
		http.Handle("/history", srvwrapper.New(handHistory.Handle))
		err = http.ListenAndServe(config.AppConfig.HttpPort, nil)
		if err != nil {
			log.Fatalln("ERR: ", err)
		}
	}()

	return &App{
		model: model,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	msgCh, err := a.model.Consume(ctx)
	if err != nil {
		return err
	}

	for msg := range msgCh {
		err = a.model.Send(msg)
		if err != nil {
			log.Printf("Error sending message: %s\n", err.Error())
		}
	}

	return nil
}
