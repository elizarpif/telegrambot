package nonsense

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/elizarpif/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

type Bot struct {
	bot *tgbotapi.BotAPI

	usersInGame map[uuid.UUID][]*userStory
	users       map[int64]uuid.UUID //чтобы понимать у кого какая игра

	waitUidUsers map[int64]struct{}
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (string, error) {
	s, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(s)), err
}

func New() (*Bot, error) {
	token, err := tokenFromFile("bot_token")
	if err != nil {
		return nil, err
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	bot.Debug = true

	return &Bot{bot: bot,
			usersInGame:  make(map[uuid.UUID][]*userStory),
			users:        make(map[int64]uuid.UUID),
			waitUidUsers: make(map[int64]struct{}),
		}, // chatID: int64(chatID)
		nil
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

		if update.Message.IsCommand() { // ignore any non-command Messages
			b.command(update)
			continue
		}

		b.processMessage(update)
	}
}

func (t *Bot) sendMessage(chatID int64, msg string) {
	botMsg := tgbotapi.NewMessage(chatID, msg)

	_, err := t.bot.Send(botMsg)
	if err != nil {
		logger.GetLogger(context.Background()).Errorf("can't send msg, err: %v", err)
	}
}

func (b *Bot) processMessage(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	if uid, ok := b.users[chatID]; ok {
		if users, ok := b.usersInGame[uid]; ok {
			for i := range users {
				if users[i].user == chatID {
					users[i].parts[users[i].partIndex] = update.Message.Text
					users[i].partIndex = users[i].partIndex + 1

					logger.GetLogger(context.Background()).Infof("%v", users[i].user)
				}
			}
		}
	}

	if _, ok := b.waitUidUsers[chatID]; ok {
		uid, err := uuid.Parse(update.Message.Text)
		if err != nil {
			b.sendMessage(chatID, "Это не правильный код, код должен быть в формате uuid, например, \"1e4e4171-f1e1-4a41-a637-a11b13d10175\"")
		} else {
			user := newUserStory(chatID, update.Message.Chat.UserName)

			b.usersInGame[uid] = append(b.usersInGame[uid], user)
			b.users[chatID] = uid

			b.sendMessage(chatID, "Вы успешно подключились к игре!")
		}

		delete(b.waitUidUsers, chatID)
	}
}

func (t *Bot) getUserNames(uid uuid.UUID) []string {
	var res []string

	for _, v := range t.usersInGame[uid] {
		res = append(res, v.username)
	}

	return res
}

func (t *Bot) command(update tgbotapi.Update) {
	// Create a new MessageConfig. We don't have text yet,
	// so we leave it empty.
	chatID := update.Message.Chat.ID
	msg := tgbotapi.NewMessage(chatID, "")

	// Extract the command from the Message.
	switch update.Message.Command() {
	case "enter_game_code":
		msg.Text = "Напишите ваш код игры"
		t.waitUidUsers[chatID] = struct{}{}

	case "generate_game_code":
		uid := uuid.New()
		msg.Text = fmt.Sprintf("Ваш код игры: %s"+
			"\nСкопируйте этот id и перешлите своим игрокам, чтобы попасть в одну игру", uid.String())

		user := newUserStory(chatID, update.Message.Chat.UserName)

		t.usersInGame[uid] = []*userStory{user}
		t.users[chatID] = uid
	case "start_game":
		uid, ok := t.users[chatID]
		if ok {
			msg.Text = fmt.Sprintf("Отлично! Мы начинаем игру ЧЕПУХА!\nТекущее кол-во игроков: %d\n%v", len(t.usersInGame[uid]), t.getUserNames(uid))

			defer func() {
				go t.gameInProcess(uid) // запустить игру в конце
			}()
		} else {
			msg.Text = "К сожалению, вы не участвуете ни в какой игре. Сгенерируйте свой код игры или присоединитесь к игре"
		}
	case "status":
		uid, ok := t.users[chatID]
		if ok {
			msg.Text = fmt.Sprintf("Текущее кол-во игроков: %d\n%v", len(t.usersInGame[uid]), t.getUserNames(uid))
		} else {
			msg.Text = "К сожалению, вы не участвуете ни в какой игре. Сгенерируйте свой код игры или присоединитесь к игре"
		}
	case "restart":
		uid, ok := t.users[chatID]
		if ok {
			for _, v := range t.usersInGame[uid] {
				v.clear()
			}

			msg.Text = fmt.Sprintf("Отлично! Мы начинаем игру ЧЕПУХА!\nТекущее кол-во игроков: %d", len(t.usersInGame[uid]))

			defer func() {
				go t.gameInProcess(uid) // запустить игру в конце
			}()
		}
	case "help", "start":
		msg.Text = fmt.Sprintf("Добро пожаловать в бот для игры \"Чепуха\"! Правила просты:\n" +
			"Зарегиструй свою игру: для этого воспользуйся командой \"generate_game_code\"\n" +
			"Если же у кого-то из игроков уже есть код, то воспользуйся командой \"enter_game_code\", чтобы присоединиться к игре\n" +
			"Узнать количество игроков по текущему коду: введи команду \"status\"\n" +
			"Если все зашли, то использую команду \"start_game\"\n" +
			"Когда игра закончена, если хотите продолжить с текущим кол-вом участников, жмякай \"restart\"")
	default:
		msg.Text = "I don't know that command"
	}

	if _, err := t.bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func (t *Bot) gameInProcess(uid uuid.UUID) {
	usersInGame, _ := t.usersInGame[uid]

	questionNum := 0
	for {

		allReady := true

		lastInd := questionNum
		for _, user := range usersInGame {
			if user.partIndex != lastInd {
				allReady = false
				break
			}
		}

		if allReady {
			if questionNum == partsNum {
				break
			}

			for _, u := range usersInGame {
				t.sendMessage(u.user, questions[questionNum])
			}

			questionNum++
		}

		time.Sleep(time.Second)
	}

	stories := generateStories(usersInGame)

	for i, s := range stories {
		for _, u := range usersInGame {
			t.sendMessage(u.user, fmt.Sprintf("%d-ая история: %s", i+1, s))
		}
	}
}
