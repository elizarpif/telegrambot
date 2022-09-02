package main

import (
	"context"
	"github.com/elizarpif/logger"
	"telegrambot/gmail"
	"telegrambot/telegrambot"
)

func main() {
	log := logger.NewLogger()
	ctx := logger.WithLogger(context.Background(), log)

	srv, err := gmail.NewService(ctx)
	if err != nil {
		log.Fatal(err)
	}

	bot, err := telegrambot.New(srv)
	if err != nil {
		log.Fatal(err)
	}

	go bot.Start(ctx)
	bot.StartTicker(ctx)
}
