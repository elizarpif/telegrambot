package main

import (
	"context"

	"github.com/elizarpif/logger"
	"github.com/elizarpif/telegrambot/telegrambot/nonsense"
)

func main() {
	log := logger.NewLogger()
	ctx := logger.WithLogger(context.Background(), log)

	bot, err := nonsense.New()
	if err != nil {
		log.Fatal(err)
	}

	bot.Start(ctx)
}
