package controllers

import (
	"net/http"

	"app/internal/controllers/gen"
	"app/internal/domain"
	"app/internal/usecase/user_usecase"

	"github.com/gin-gonic/gin"
)

type UserController interface {
	PostUsersSetIsActive(c *gin.Context)
	PostUsersDeactivateTeam(c *gin.Context)
}

type userController struct {
	userUseCase user_usecase.UserUseCase
}

func (s *userController) PostUsersDeactivateTeam(c *gin.Context) {
	var req gen.PostUsersDeactivateTeamJSONBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := s.userUseCase.DeactivateUsersByTeamName(c.Request.Context(), req.TeamName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Users deactivated successfully"})
}

func NewUserController(userUseCase user_usecase.UserUseCase) UserController {
	return &userController{
		userUseCase: userUseCase,
	}
}

func (s *userController) PostUsersSetIsActive(c *gin.Context) {
	var req gen.PostUsersSetIsActiveJSONBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.userUseCase.UpdateUserActivity(c.Request.Context(), domain.UserID(req.UserId), domain.UserActivityStatus(req.IsActive))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": user})
}
