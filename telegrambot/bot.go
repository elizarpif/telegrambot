package telegrambot

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/elizarpif/logger"
	"github.com/elizarpif/telegrambot/gmail"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	gmailSrv gmail.GetFiverrMessages
	bot      *tgbotapi.BotAPI

	chatID int64
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (string, error) {
	s, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(s)), err
}

func New(gmailSrv gmail.GetFiverrMessages) (*Bot, error) {
	token, err := tokenFromFile("bot_token")
	if err != nil {
		return nil, err
	}

	chatIDStr, err := tokenFromFile("chat_id")
	if err != nil {
		return nil, err
	}
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		return nil, err
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	bot.Debug = true
	return &Bot{bot: bot, gmailSrv: gmailSrv, chatID: int64(chatID)}, nil
}

func (b *Bot) fiverrMsg(ctx context.Context) {
	msgs := b.gmailSrv.GetNewFiverrMsg(ctx)

	for _, msg := range msgs {
		botMsg := tgbotapi.NewMessage(b.chatID, fmt.Sprintf("New message!\n%s", msg))
		sentMsg, err := b.bot.Send(botMsg)
		if err != nil {
			logger.GetLogger(ctx).Errorf("can't send msg, err: %v", err)
		}

		edit := tgbotapi.EditMessageTextConfig{
			BaseEdit: tgbotapi.BaseEdit{
				ChatID:    b.chatID,
				MessageID: sentMsg.MessageID,
			},
			Text:      fmt.Sprintf("New message!\n%s", msg),
			ParseMode: "HTML",
		}
		if _, err := b.bot.Send(edit); err != nil {
			logger.GetLogger(ctx).Errorf("can't send msg, err: %v", err)
		}
	}
}

func (b *Bot) StartTicker(ctx context.Context) {
	b.fiverrMsg(ctx)

	ticker := time.NewTicker(time.Minute * 20)

	for range ticker.C {
		b.fiverrMsg(ctx)
	}
}

func (b *Bot) Start(ctx context.Context) {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	// Start polling Telegram for updates.
	updates := b.bot.GetUpdatesChan(updateConfig)

	// Let's go through each update that we're getting from Telegram.
	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

		msg.ReplyToMessageID = update.Message.MessageID

		if _, err := b.bot.Send(msg); err != nil {
			logger.GetLogger(ctx).Printf("can't send msg\n")
		}
	}

}
