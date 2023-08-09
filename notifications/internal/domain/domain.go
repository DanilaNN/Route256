package domain

import (
	"context"
	"fmt"
	"log"
	"route256/notifications/internal/consumer"
	"route256/notifications/internal/domain/models"
	"time"
)

const ServiceName = "Notifications"

type CacheKey struct {
	User   int64
	Year   int
	Month  int
	Day    int
	Hour   int
	Minute int
}

type Model struct {
	sender   Sender
	consumer Consumer
	cache    Cache
}

func New(sender Sender, consumer Consumer, cache Cache) *Model {
	return &Model{
		sender:   sender,
		consumer: consumer,
		cache:    cache,
	}
}

type Sender interface {
	SendMessage(order models.Order) error
}

type Consumer interface {
	Listen(ctx context.Context)
	GetResultChannel() <-chan consumer.ConsumerMessage
	Close() error
}

type Cache interface {
	Add(key CacheKey, value interface{}) bool
	Get(key CacheKey) interface{}
	Remove(key CacheKey) bool
	Len() int
}

func (m *Model) Send(order models.Order) error {
	err := m.sender.SendMessage(order)
	if err != nil {
		return err
	}

	return nil
}

func (m *Model) Consume(ctx context.Context) (<-chan models.Order, error) {

	go func() {
		defer func() {
			err := m.consumer.Close()
			if err != nil {
				log.Printf("consumer close %s\n", err.Error())
			}
		}()
		m.consumer.Listen(ctx)
	}()

	outCh := make(chan models.Order)

	go func(ctx context.Context) {
		defer close(outCh)
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-m.consumer.GetResultChannel():

				m.AddToCache(CacheKey{
					User:   msg.User,
					Year:   msg.TimeStamp.Year(),
					Month:  int(msg.TimeStamp.Month()),
					Day:    msg.TimeStamp.Day(),
					Hour:   msg.TimeStamp.Hour(),
					Minute: msg.TimeStamp.Minute(),
				}, msg.Status)

				outCh <- models.Order{
					User: msg.User, OrderId: uint64(msg.OrderID), Status: msg.Status}
			}
		}
	}(ctx)

	return outCh, nil
}

func (m *Model) AddToCache(key CacheKey, status string) {

	// check if slice with statuses for current minute already exist
	// if exist, append new status
	statuses := m.cache.Get(key)
	if statuses != nil {
		sliceVal := statuses.([]string)
		sliceVal = append(sliceVal, status)
		m.cache.Add(key, sliceVal)
		return
	}

	m.cache.Add(key, []string{status})
}

func (m *Model) GetFromCache(user int64, timeStart, timeStop time.Time) ([]string, error) {

	fmt.Println(m.cache.Len())

	out := make([]string, 0)
	for timeStart.Before(timeStop) {
		statuses := m.cache.Get(CacheKey{
			User:   user,
			Year:   timeStart.Year(),
			Month:  int(timeStart.Month()),
			Day:    timeStart.Day(),
			Hour:   timeStart.Hour(),
			Minute: timeStart.Minute(),
		})
		if statuses != nil {
			sliceVal := statuses.([]string)
			out = append(out, sliceVal...)
		}

		timeStart = timeStart.Add(1 * time.Minute)
	}

	if len(out) == 0 {
		return out, fmt.Errorf("cache miss")
	}

	return out, nil
}
