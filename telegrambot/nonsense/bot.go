package nonsense

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/elizarpif/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

type Bot struct {
	bot *tgbotapi.BotAPI

	usersInGame map[uuid.UUID]*game
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
			usersInGame:  make(map[uuid.UUID]*game),
			users:        make(map[int64]uuid.UUID),
			waitUidUsers: make(map[int64]struct{}),
		},
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
			b.command(ctx, update)
			continue
		}

		b.processMessage(ctx, update)
	}
}

func (t *Bot) sendMessage(ctx context.Context, chatID int64, msg string) {
	botMsg := tgbotapi.NewMessage(chatID, msg)

	_, err := t.bot.Send(botMsg)
	if err != nil {
		logger.GetLogger(ctx).Errorf("can't send msg, err: %v", err)
	}
}

func (b *Bot) processMessage(ctx context.Context, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	if uid, ok := b.users[chatID]; ok {
		if game, ok := b.usersInGame[uid]; ok {
			for i := range game.users {
				if game.users[i].user == chatID {
					game.users[i].parts[game.partIndex] = update.Message.Text // todo ++partindex in processgame
					game.updatedAt = time.Now()
				}
			}
		}
	}

	if _, ok := b.waitUidUsers[chatID]; ok {
		uid, err := uuid.Parse(update.Message.Text)
		if err != nil {
			b.sendMessage(ctx, chatID, "Это неправильный код, код должен быть в формате uuid, например, \"1e4e4171-f1e1-4a41-a637-a11b13d10175\"")
		} else {
			user := newUserStory(chatID, update.Message.Chat.UserName)

			b.usersInGame[uid].users = append(b.usersInGame[uid].users, user)
			b.users[chatID] = uid

			b.sendMessage(ctx, chatID, "Вы успешно подключились к игре!")
		}

		delete(b.waitUidUsers, chatID)
	}
}

func (t *Bot) getUserNames(uid uuid.UUID) []string {
	var res []string

	for _, v := range t.usersInGame[uid].users {
		res = append(res, v.username)
	}

	return res
}

func (t *Bot) command(ctx context.Context, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	text := ""

	switch update.Message.Command() {
	case "enter_game_code":
		text = "Напишите ваш код игры"
		t.waitUidUsers[chatID] = struct{}{}

	case "generate_game_code":
		uid := uuid.New()
		text = fmt.Sprintf("Ваш код игры: %s"+
			"\nСкопируйте этот id и перешлите своим игрокам, чтобы попасть в одну игру", uid.String())

		user := newUserStory(chatID, update.Message.Chat.UserName)

		t.usersInGame[uid] = &game{users: []*userStory{user}}
		t.users[chatID] = uid
	case "start_game":
		uid, ok := t.users[chatID]
		if ok {
			text = fmt.Sprintf("Отлично! Мы начинаем игру ЧЕПУХА!\nТекущее кол-во игроков: %d\n%v",
				len(t.usersInGame[uid].users), t.getUserNames(uid))

			defer func() {
				go t.gameInProcess(ctx, uid) // запустить игру в конце
			}()
		} else {
			text = "К сожалению, вы не участвуете ни в какой игре. Сгенерируйте свой код игры или присоединитесь к игре"
		}
	case "status":
		uid, ok := t.users[chatID]
		if ok {
			text = fmt.Sprintf("Текущее кол-во игроков: %d\n%v", len(t.usersInGame[uid].users), t.getUserNames(uid))
		} else {
			text = "К сожалению, вы не участвуете ни в какой игре. Сгенерируйте свой код игры или присоединитесь к игре"
		}
	case "restart":
		uid, ok := t.users[chatID]
		if ok {
			for _, v := range t.usersInGame[uid].users {
				v.clear()
			}

			text = fmt.Sprintf("Отлично! Мы начинаем игру ЧЕПУХА!\nТекущее кол-во игроков: %d", len(t.usersInGame[uid].users))

			defer func() {
				go t.gameInProcess(ctx, uid) // запустить игру в конце
			}()
		}
	case "help", "start":
		text = fmt.Sprintf(`Добро пожаловать в бот для игры "Чепуха"! Правила просты:
			Зарегиструй свою игру: для этого воспользуйся командой generate_game_code
			Если же у кого-то из игроков уже есть код, то воспользуйся командой enter_game_code, чтобы присоединиться к игре
			Узнать количество игроков по текущему коду: введи команду status
			Если все зашли, то использую команду start_game
			Когда игра закончена, если хотите продолжить с текущим кол-вом участников, жмякай restart`)
	default:
		text = "I don't know that command"
	}

	t.sendMessage(ctx, chatID, text)
}

// todo переписать этот жуткий алгоритм, написанный посередь ночи
func (t *Bot) gameInProcess(ctx context.Context, uid uuid.UUID) {
	usersInGame, _ := t.usersInGame[uid]

	questionNum := 0
	for _, u := range usersInGame.users {
		t.sendMessage(ctx, u.user, questions[0])
	}

	for questionNum < partsNum {
		allReady := true

		for _, v := range usersInGame.users {
			//  у всех юзеров должны быть не пустые поля
			if v.parts[questionNum] == "" {
				allReady = false
				break
			}
		}
		if allReady {
			questionNum++
			if questionNum == partsNum {
				break
			}
			for _, u := range usersInGame.users {
				t.sendMessage(ctx, u.user, questions[questionNum])
			}

			usersInGame.partIndex++
		}

		time.Sleep(time.Second)
	}

	stories := generateStories(usersInGame.users)

	for i, s := range stories {
		for _, u := range usersInGame.users {
			t.sendMessage(ctx, u.user, fmt.Sprintf("%d-ая история: \n%s", i+1, s))
		}
	}
}
