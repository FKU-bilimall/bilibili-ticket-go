package token

import "time"

type Generator interface {
	GenerateTokenPrepareStage() string
	GenerateTokenCreateStage(whenGenPToken time.Time) string
	IsHotProject() bool
}
