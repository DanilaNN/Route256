package sender

import (
	"fmt"
	"route256/notifications/internal/domain/models"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/pkg/errors"
)

type Sender struct {
	bot    *tgbotapi.BotAPI
	chatId uint64
}

func NewTelegramSender(bot *tgbotapi.BotAPI, chatId uint64) *Sender {
	return &Sender{
		bot:    bot,
		chatId: chatId,
	}
}

func (s *Sender) SendMessage(order models.Order) error {
	text := fmt.Sprintf("order_id = %d, order_status = %s\n", order.OrderId, order.Status)

	msg := tgbotapi.NewMessage(int64(s.chatId), text)

	if _, err := s.bot.Send(msg); err != nil {
		return errors.Wrap(err, "telegram message send")
	}

	return nil
}
