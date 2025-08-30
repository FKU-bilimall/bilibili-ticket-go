package token

import "time"

type NormalTokenGenerator struct{}

func NewNormalTokenGenerator() *NormalTokenGenerator {
	return &NormalTokenGenerator{}
}

func (g *NormalTokenGenerator) GenerateTokenPrepareStage() string {
	return ""
}

func (g *NormalTokenGenerator) GenerateTokenCreateStage(_ time.Time) string {
	return ""
}

func (g *NormalTokenGenerator) IsHotProject() bool {
	return false
}
