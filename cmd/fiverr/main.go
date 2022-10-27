package main

import (
	"context"
	"github.com/elizarpif/telegrambot/telegrambot/fiverr"

	"github.com/elizarpif/logger"
	"github.com/elizarpif/telegrambot/gmail"
)

func main() {
	log := logger.NewLogger()
	ctx := logger.WithLogger(context.Background(), log)

	srv, err := gmail.NewService(ctx)
	if err != nil {
		log.Fatal(err)
	}

	bot, err := fiverr.New(srv)
	if err != nil {
		log.Fatal(err)
	}

	go bot.Start(ctx)
	bot.StartTicker(ctx)
}
