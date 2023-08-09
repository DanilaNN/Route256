package kafka

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Shopify/sarama"
)

type OrderMessage struct {
	User    int64
	OrderID uint64
	Status  string
}

type ConsumerGroup struct {
	ready    chan bool
	Messages chan *sarama.ConsumerMessage
}

func NewConsumerGroup() ConsumerGroup {
	return ConsumerGroup{
		ready:    make(chan bool),
		Messages: make(chan *sarama.ConsumerMessage),
	}
}

func (consumer *ConsumerGroup) Ready() <-chan bool {
	return consumer.ready
}

// Setup Начинаем новую сессию, до ConsumeClaim
func (consumer *ConsumerGroup) Setup(_ sarama.ConsumerGroupSession) error {
	close(consumer.ready)

	return nil
}

// Cleanup завершает сессию, после того, как все ConsumeClaim завершатся
func (consumer *ConsumerGroup) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim читаем до тех пор пока сессия не завершилась
func (consumer *ConsumerGroup) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():

			pm := OrderMessage{}
			err := json.Unmarshal(message.Value, &pm)
			if err != nil {
				fmt.Println("Consumer group error", err)
			}

			log.Printf("Message claimed: value = %v, timestamp = %v, topic = %s",
				pm,
				message.Timestamp,
				message.Topic,
			)

			consumer.Messages <- message

			// коммит сообщения "руками"
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}
