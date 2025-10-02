package notify

import (
	"net/url"

	"github.com/imroc/req/v3"
)

type Gotify struct {
	token    string
	endpoint string
}

func NewGotify(token, endpoint string) *Gotify {
	return &Gotify{
		token:    token,
		endpoint: endpoint,
	}
}

func (g *Gotify) Notify(msg string) bool {
	result, err := url.JoinPath(g.endpoint, "message")
	if err != nil {
		logger.Warn(err)
		return false
	}
	res, err := req.R().SetHeader("Authorization", "Bearer "+g.token).SetBodyJsonMarshal(map[string]any{
		"message":  msg,
		"priority": 8,
		"title":    "Bili-Ticket-Go",
	}).Post(result)
	if err != nil {
		logger.Warn(err)
		return false
	} else if res.IsErrorState() {
		return false
	}
	return true
}

func (g *Gotify) Test() bool {
	return g.Notify("This is a test message from Bili-Ticket-Go.")
}
