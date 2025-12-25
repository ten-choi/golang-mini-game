package graph

import (
	"draw-and-guess-server/internal/service"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	UserService        service.UserService
	QuizService        service.QuizService
	PlayerStatsService service.PlayerStatsService
}
