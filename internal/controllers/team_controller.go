package controllers

import (
	"net/http"

	"app/internal/controllers/gen"
	"app/internal/domain"
	"app/internal/mapper"
	"app/internal/usecase/team_usecase"

	"github.com/gin-gonic/gin"
)

type TeamController interface {
	PostTeamAdd(c *gin.Context)
	GetTeamGet(c *gin.Context, params gen.GetTeamGetParams)
}

type teamController struct {
	teamUseCase team_usecase.TeamUseCase
}

func NewTeamController(teamUseCase team_usecase.TeamUseCase) TeamController {
	return &teamController{
		teamUseCase: teamUseCase,
	}
}

func (s *teamController) GetTeamGet(c *gin.Context, params gen.GetTeamGetParams) {
	team, err := s.teamUseCase.GetTeamByName(c.Request.Context(), params.TeamName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}



	c.JSON(http.StatusOK, gin.H{"team": mapper.DomainTeamToDTO(*team)})
}

func (s *teamController) PostTeamAdd(c *gin.Context) {
	var req gen.PostTeamAddJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var users []domain.TeamUser
	for _, member := range req.Members {
		users = append(users, domain.TeamUser{
			ID: 	domain.UserID(member.UserId),
			Name: 	member.Username,
		})
	}

	team, err := s.teamUseCase.CreateTeam(c.Request.Context(), req.TeamName, users)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"team": mapper.DomainTeamToDTO(*team)})	
}
