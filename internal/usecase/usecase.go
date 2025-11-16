package usecase

import (
	"app/internal/usecase/pr_usecase"
	"app/internal/usecase/stats_usecase"
	"app/internal/usecase/team_usecase"
	"app/internal/usecase/user_usecase"
)


type UseCase interface {
	user_usecase.UserUseCase
	team_usecase.TeamUseCase
	stats_usecase.StatsUseCase
	pr_usecase.PullRequestUseCase
}


type useCase struct {
	user_usecase.UserUseCase
	team_usecase.TeamUseCase
	pr_usecase.PullRequestUseCase
	stats_usecase.StatsUseCase
}

func NewUseCase(
	userUseCase user_usecase.UserUseCase,
	teamUseCase team_usecase.TeamUseCase,
	prUseCase pr_usecase.PullRequestUseCase,
	statsUseCase stats_usecase.StatsUseCase,
) UseCase {
	return &useCase{
		UserUseCase:        userUseCase,
		TeamUseCase:        teamUseCase,
		PullRequestUseCase: prUseCase,
		StatsUseCase:       statsUseCase,
	}
}