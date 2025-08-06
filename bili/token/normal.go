package token

import (
	"time"
)

type NormalToken struct{}

func NewNormalTokenGenerator() *NormalToken {
	return &NormalToken{}
}

func (g *NormalToken) GenerateTokenPrepareStage() string {
	return ""
}

func (g *NormalToken) GenerateTokenCreateStage(whenGenPToken time.Time) string {
	return ""
}

func (g *NormalToken) IsHotProject() bool {
	return true
}
