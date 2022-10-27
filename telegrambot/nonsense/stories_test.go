package nonsense

import "testing"

func Test_oneStory(t *testing.T) {
	type args struct {
		indexUser int
		stories   []*userStory
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "positive, 2 users",
			args: args{
				indexUser: 0,
				stories: []*userStory{
					{
						chatID: 1,
						parts: [7]string{
							"олень", "с козлом", "утром", "в лесу", "играли в догонялки", "им сказали плохо", "они поженились",
						},
					},
					{
						chatID: 2,
						parts: [7]string{
							"собака", "с псом", "вечером", "в городе", "сидели в ресторане", "им ничего не сказали", "они умерли",
						},
					},
				},
			},
			want: `(Кто?) олень
(С кем?) с псом
(Где?) утром
(Когда?) в городе
(Что делали?) играли в догонялки
(Что им сказали?) им ничего не сказали
(Чем все закончилось?) они поженились
`,
		},
		{
			name: "positive, 2 users, index = 1",
			args: args{
				indexUser: 1,
				stories: []*userStory{
					{
						chatID: 1,
						parts: [7]string{
							"олень", "с козлом", "утром", "в лесу", "играли в догонялки", "им сказали плохо", "они поженились",
						},
					},
					{
						chatID: 2,
						parts: [7]string{
							"собака", "с псом", "вечером", "в городе", "сидели в ресторане", "им ничего не сказали", "они умерли",
						},
					},
				},
			},
			want: `(Кто?) собака
(С кем?) с козлом
(Где?) вечером
(Когда?) в лесу
(Что делали?) сидели в ресторане
(Что им сказали?) им сказали плохо
(Чем все закончилось?) они умерли
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := oneStory(tt.args.indexUser, tt.args.stories); got != tt.want {
				t.Errorf("oneStory() = %v, want %v", got, tt.want)
			}
		})
	}
}
