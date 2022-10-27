package nonsense

import "strings"

const partsNum = 7

type userStory struct {
	user      int64
	username  string
	parts     [partsNum]string
	partIndex int
}

func newUserStory(user int64, username string) *userStory {
	return &userStory{user: user, username: username}
}

func (u *userStory) clear() {
	u.partIndex = 0
	u.parts = [7]string{}
}

var questions = [7]string{
	"Кто?", "С кем?", "где?",
	"Когда?",
	"Что делали?",
	"Что им сказали?",
	"Чем все закончилось?",
}

func generateStories(userStories []*userStory) []string {
	usersCount := len(userStories)
	stories := make([]string, 0, usersCount)

	for i := range userStories {
		stories = append(stories, oneStory(i, userStories))
	}

	return stories
}

func oneStory(indexUser int, stories []*userStory) string {
	if indexUser >= len(stories) {
		return ""
	}

	b := strings.Builder{}

	for i := 0; i < partsNum; i++ {
		ind := (indexUser + i) % len(stories)

		_, err := b.WriteString(stories[ind].parts[i])

		if err != nil {
			return ""
		}

		_, err = b.WriteString(" ")

		if err != nil {
			return ""
		}
	}

	return b.String()
}

/*
команда: войти в игру (сгенерировать свой код или же войти по коду)
команда: готов начать
- игра раунд 1 начата! -> получить ответы от всех участников
- кто?
- с кем?
- где?
- когда?
- что делали?
- что им сказала?
- чем все закончилось?
-----результаты----
listUsers = []int64{}
responses = [len(users)][len(questions)]
result:
Как присоединиться к игре по коду?

*/
