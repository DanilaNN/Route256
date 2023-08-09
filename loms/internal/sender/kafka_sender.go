package sender

import (
	"encoding/json"
	"fmt"
	"route256/loms/internal/infrastructure/kafka"

	"github.com/Shopify/sarama"
)

type PaymentMessage struct {
	User    int64
	OrderID uint64
	Status  string
}

type KafkaSender struct {
	producer *kafka.Producer
	topic    string
}

func New(producer *kafka.Producer, topic string) *KafkaSender {
	return &KafkaSender{
		producer,
		topic,
	}
}

func (s *KafkaSender) SendMessage(message PaymentMessage) error {
	kafkaMsg, err := s.buildMessage(message)
	if err != nil {
		fmt.Println("Send message marshal error", err)
		return err
	}

	partition, offset, err := s.producer.SendSyncMessage(kafkaMsg)

	if err != nil {
		fmt.Println("Send message connector error", err)
		return err
	}

	fmt.Println("Partition: ", partition, " Offset: ", offset, " AnswerID:", message.OrderID)
	return nil
}

func (s *KafkaSender) SendMessages(messages []PaymentMessage) error {
	var kafkaMsg []*sarama.ProducerMessage
	var message *sarama.ProducerMessage
	var err error

	for _, m := range messages {
		message, err = s.buildMessage(m)
		kafkaMsg = append(kafkaMsg, message)

		if err != nil {
			fmt.Println("Send message marshal error", err)
			return err
		}
	}

	err = s.producer.SendSyncMessages(kafkaMsg)

	if err != nil {
		fmt.Println("Send message connector error", err)
		return err
	}

	fmt.Println("Send messages count:", len(messages))
	return nil
}

func (s *KafkaSender) buildMessage(message PaymentMessage) (*sarama.ProducerMessage, error) {
	msg, err := json.Marshal(message)

	if err != nil {
		fmt.Println("Send message marshal error", err)
		return nil, err
	}

	return &sarama.ProducerMessage{
		Topic:     s.topic,
		Value:     sarama.ByteEncoder(msg),
		Partition: -1,
		Key:       sarama.StringEncoder(fmt.Sprint(message.OrderID)),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("test-header"),
				Value: []byte("test-value"),
			},
		},
	}, nil
}
