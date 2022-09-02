package gmail

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type GetFiverrMessages interface {
	GetNewFiverrMsg(ctx context.Context) []string
}

type Service struct {
	srv *gmail.Service
}

func NewService(ctx context.Context) (*Service, error) {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Gmail client: %v", err)
	}

	return &Service{srv: srv}, nil
}

func (gs *Service) GetNewFiverrMsg(ctx context.Context) []string {
	user := "me"

	list := gs.srv.Users.Threads.List(user)
	list.Q("from:noreply@e.fiverr.com in:unread")

	resp, err := list.Do()
	if err != nil {
		fmt.Println("No messages")
		return nil
	}

	if len(resp.Threads) == 0 {
		return nil
	}

	res := make([]string, 0, len(resp.Threads))

	for _, l := range resp.Threads {
		res = append(res, l.Snippet)
	}

	return res
}
