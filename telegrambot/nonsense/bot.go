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

	usersInGameSync  *usersInGameNap
	usersSync        *usersMap //чтобы понимать у кого какая игра
	waitUidUsersSync *waitUidUsers
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

	return &Bot{
			bot: bot,

			usersInGameSync:  newUsersInGameNap(),
			usersSync:        newUsersMap(),
			waitUidUsersSync: newWaitUidUsers(),
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

func (b *Bot) sendMessage(ctx context.Context, chatID int64, msg string) {
	botMsg := tgbotapi.NewMessage(chatID, msg)

	_, err := b.bot.Send(botMsg)
	if err != nil {
		logger.GetLogger(ctx).Errorf("can't send msg, err: %v", err)
	}
}

func (b *Bot) processMessage(ctx context.Context, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	// юзеры которые присылают код для присоединения к игре
	if _, ok := b.waitUidUsersSync.Load(chatID); ok {
		uid, err := uuid.Parse(update.Message.Text)
		if err != nil {
			b.sendMessage(ctx, chatID, "Это неправильный код, код должен быть в формате uuid, например, \"1e4e4171-f1e1-4a41-a637-a11b13d10175\"")
		} else {
			user := newUserStory(chatID, update.Message.Chat.UserName)

			if game, ok := b.usersInGameSync.Load(uid); ok {
				if !ok {
					logger.GetLogger(ctx).Errorf("can't cast to userStory")
				} else {
					game.users = append(game.users, user)

					b.usersSync.Store(chatID, uid)
					b.usersInGameSync.Store(uid, game) // store?

					b.sendMessage(ctx, chatID, "Вы успешно подключились к игре!")
				}
			}
		}

		b.waitUidUsersSync.Delete(chatID)
		return
	}

	// юзеры которые находятся в игре
	if uid, ok := b.usersSync.Load(chatID); ok {
		if game, ok := b.usersInGameSync.Load(uid); ok {
			if !ok {
				logger.GetLogger(ctx).Errorf("can't cast to userStory")
			} else {
				for i := range game.users {
					if game.users[i].chatID == chatID {
						game.users[i].parts[game.partIndex] = update.Message.Text // todo ++partindex in processgame
						game.updatedAt = time.Now()
					}
				}
				b.usersInGameSync.Store(uid, game)
			}
		}
	}
}

func (b *Bot) getUserNames(uid uuid.UUID) []string {
	var res []string

	if game, ok := b.usersInGameSync.Load(uid); ok {
		for _, v := range game.users {
			res = append(res, v.username)
		}
	}

	return res
}

func (b *Bot) command(ctx context.Context, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	text := ""

	switch update.Message.Command() {
	case "enter_game_code":
		text = "Напишите ваш код игры"
		b.waitUidUsersSync.Store(chatID)

	case "generate_game_code":
		uid := uuid.New()
		text = fmt.Sprintf("Ваш код игры: %s"+
			"\nСкопируйте этот id и перешлите своим игрокам, чтобы попасть в одну игру", uid.String())
		user := newUserStory(chatID, update.Message.Chat.UserName)

		b.usersInGameSync.Store(uid, &game{users: []*userStory{user}})
		b.usersSync.Store(chatID, uid)

	case "start_game":
		uid, ok := b.usersSync.Load(chatID)
		if ok {
			game, ok := b.usersInGameSync.Load(uid)
			if !ok {
				text = "К сожалению, вы не участвуете ни в какой игре. Сгенерируйте свой код игры или присоединитесь к игре"
				break
			}
			if game.inProcess {
				text = "Вы уже участвуете в игре"
				break
			}
			game.clear()
			b.usersInGameSync.Store(uid, game)

			text = fmt.Sprintf("Отлично! Мы начинаем игру ЧЕПУХА!\nТекущее кол-во игроков: %d\n%v",
				len(game.users), b.getUserNames(uid))

			defer func() {
				go b.gameInProcess(ctx, uid) // запустить игру в конце
			}()
		} else {
			text = "К сожалению, вы не участвуете ни в какой игре. Сгенерируйте свой код игры или присоединитесь к игре"
		}
	case "status":
		uid, ok := b.usersSync.Load(chatID)
		if ok {
			game, _ := b.usersInGameSync.Load(uid)
			text = fmt.Sprintf("Текущее кол-во игроков: %d\n%v", len(game.users), b.getUserNames(uid))
		} else {
			text = "К сожалению, вы не участвуете ни в какой игре. Сгенерируйте свой код игры или присоединитесь к игре"
		}
	case "restart":
		uid, ok := b.usersSync.Load(chatID)
		if ok {
			game, ok := b.usersInGameSync.Load(uid)
			if !ok {
				text = "К сожалению, вы не участвуете ни в какой игре. Сгенерируйте свой код игры или присоединитесь к игре"
				break
			}
			if game.partIndex != partsNum {
				text = fmt.Sprintf("Кажется, вы еще не закончили игру, текущий этап: %v", game.partIndex)
				break
			}

			game.clear()
			b.usersInGameSync.Store(uid, game)

			text = fmt.Sprintf("Отлично! Мы начинаем игру ЧЕПУХА!\nТекущее кол-во игроков: %d", len(game.users))

			defer func() {
				go b.gameInProcess(ctx, uid) // запустить игру в конце
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

	b.sendMessage(ctx, chatID, text)
}

// todo переписать этот жуткий алгоритм, написанный посередь ночи
func (b *Bot) gameInProcess(ctx context.Context, uid uuid.UUID) {
	usersInGame, _ := b.usersInGameSync.Load(uid)

	usersInGame.inProcess = true
	questionNum := 0

	for _, u := range usersInGame.users {
		b.sendMessage(ctx, u.chatID, questions[0])
	}

	for {
		allReady := true

		for _, v := range usersInGame.users {
			if userUid, ok := b.usersSync.Load(v.chatID); ok && uid != userUid {
				for _, u := range usersInGame.users {
					b.sendMessage(ctx, u.chatID, fmt.Sprintf("участник %s покинул игру %s и теперь в игре %s. Игра закончена", u.username, uid, userUid))
				}
				logger.GetLogger(ctx).Errorf("user left the game")
				return
			}
			//  у всех юзеров должны быть не пустые поля
			if v.parts[questionNum] == "" {
				allReady = false
				break
			}
		}
		if allReady {
			usersInGame.partIndex++

			questionNum++

			if questionNum == partsNum {
				break
			}
			for _, u := range usersInGame.users {
				b.sendMessage(ctx, u.chatID, questions[questionNum])
			}
		}

		time.Sleep(time.Second)
	}

	usersInGame.inProcess = false
	// перемешать ответы в истории
	stories := generateStories(usersInGame.users)

	// разослать всем юзерам результат
	for i, s := range stories {
		for _, u := range usersInGame.users {
			b.sendMessage(ctx, u.chatID, fmt.Sprintf("%d-ая история: \n%s", i+1, s))
		}
	}
}
