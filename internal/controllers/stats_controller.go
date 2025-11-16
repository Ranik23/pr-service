package controllers

import (
	"net/http"

	"app/internal/domain"
	"app/internal/usecase/stats_usecase"
	"app/internal/controllers/gen"

	"github.com/gin-gonic/gin"
)

type StatsController interface {
	GetStatsAssignments(c *gin.Context, params gen.GetStatsAssignmentsParams)
}

type statsController struct {
	statsUseCase stats_usecase.StatsUseCase
}

func NewStatsController(statsUseCase stats_usecase.StatsUseCase) StatsController {
	return &statsController{
		statsUseCase: statsUseCase,
	}
}

func (s *statsController) GetStatsAssignments(c *gin.Context, params gen.GetStatsAssignmentsParams) {
	userStats, err := s.statsUseCase.GetAssignCountByUserID(c.Request.Context(), domain.UserID(params.UserId))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":        userStats.UserID.String(),
		"assigned_count": userStats.AssignedCount,
	})
}	