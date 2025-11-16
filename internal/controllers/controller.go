package controllers

import (
	"app/internal/controllers/gen"
)




type Controller struct {
	UserController
	TeamController
	StatsController
	PullRequestController
}

func NewController(userController UserController, teamController TeamController,
	statsController StatsController, pullRequestController PullRequestController,) gen.ServerInterface {
	return &Controller{
		UserController:        userController,
		TeamController:        teamController,
		StatsController:       statsController,
		PullRequestController: pullRequestController,
	}
}