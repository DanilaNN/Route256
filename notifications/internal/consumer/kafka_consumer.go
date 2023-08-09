package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"route256/notifications/internal/infrastructure/kafka"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
)

type ConsumerMessage struct {
	User      int64
	OrderID   uint64
	Status    string
	TimeStamp time.Time
}

type Consumer struct {
	consumer *kafka.ConsumerGroup
	client   sarama.ConsumerGroup
	group    string
	brokers  []string
	resultCh chan ConsumerMessage
}

func New(consumer *kafka.ConsumerGroup, group string, brokers []string) (*Consumer, error) {
	/**
	 * Construct a new Sarama configuration.
	 * The Kafka cluster version has to be defined before the consumer/producer is initialized.
	 */
	config := sarama.NewConfig()
	config.Version = sarama.MaxVersion

	/*
	 sarama.OffsetNewest - получаем только новые сообщений, те, которые уже были игнорируются
	 sarama.OffsetOldest - читаем все с самого начала
	*/
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	// Используется, если ваш offset "уехал" далеко и нужно пропустить невалидные сдвиги
	config.Consumer.Group.ResetInvalidOffsets = true

	// Сердцебиение консьюмера
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second

	// Таймаут сессии
	config.Consumer.Group.Session.Timeout = 60 * time.Second

	// Таймаут ребалансировки
	config.Consumer.Group.Rebalance.Timeout = 60 * time.Second

	const BalanceStrategy = "roundrobin"
	switch BalanceStrategy {
	case "sticky":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategySticky}
	case "roundrobin":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategyRoundRobin}
	case "range":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategyRange}
	default:
		return nil, fmt.Errorf("unrecognized consumer group partition assignor: %s", BalanceStrategy)
	}

	/**
	 * Setup a new Sarama consumer group
	 */
	client, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		return nil, fmt.Errorf("error creating consumer group client: %v", err)
	}

	return &Consumer{
		consumer: consumer,
		client:   client,
		group:    group,
		brokers:  brokers,
		resultCh: make(chan ConsumerMessage),
	}, nil
}

func (c *Consumer) Close() error {
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("error closing client: %v", err)
	}
	return nil
}

func (c *Consumer) GetResultChannel() <-chan ConsumerMessage {
	return c.resultCh
}

func (c *Consumer) Listen(ctx context.Context) {

	defer close(c.resultCh)

	keepRunning := true
	log.Println("Starting a new Sarama consumer")

	consumptionIsPaused := false

	go func() {
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := c.client.Consume(ctx, []string{"orders"}, c.consumer); err != nil {
				log.Printf("Error from consumer: %v\n", err)
				time.Sleep(1 * time.Second)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
		}
	}()

	<-c.consumer.Ready() // Await till the consumer has been set up
	log.Println("Sarama consumer up and running!...")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-c.consumer.Messages:
				om := kafka.OrderMessage{}
				err := json.Unmarshal(msg.Value, &om)
				if err != nil {
					fmt.Println("Consumer group error", err)
				}
				log.Printf("Msg was sent: %v\n", om)
				c.resultCh <- ConsumerMessage{
					User:      om.User,
					OrderID:   om.OrderID,
					Status:    om.Status,
					TimeStamp: msg.Timestamp,
				}
			}
		}
	}()

	sigusr1 := make(chan os.Signal, 1)
	signal.Notify(sigusr1, syscall.SIGUSR1)

	for keepRunning {
		select {
		case <-ctx.Done():
			log.Println("terminating: context cancelled")
			keepRunning = false
		case <-sigusr1:
			log.Println("toggleConsumptionFlow")
			toggleConsumptionFlow(c.client, &consumptionIsPaused)
		}
	}
}

func toggleConsumptionFlow(client sarama.ConsumerGroup, isPaused *bool) {
	if *isPaused {
		client.ResumeAll()
		log.Println("Resuming consumption")
	} else {
		client.PauseAll()
		log.Println("Pausing consumption")
	}

	*isPaused = !*isPaused
}
