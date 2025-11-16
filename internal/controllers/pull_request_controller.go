package controllers

import (
	"net/http"

	"app/internal/controllers/gen"
	"app/internal/domain"
	"app/internal/mapper"
	"app/internal/usecase/pr_usecase"

	"github.com/gin-gonic/gin"
)

type PullRequestController interface {
	PostPullRequestCreate(c *gin.Context)
	PostPullRequestMerge(c *gin.Context)
	PostPullRequestReassign(c *gin.Context)
	GetUsersGetReview(c *gin.Context, params gen.GetUsersGetReviewParams)
}

type pullRequestController struct {
	pullRequestUseCase pr_usecase.PullRequestUseCase
}

func NewPullRequestController(pullRequestUseCase pr_usecase.PullRequestUseCase) PullRequestController {
	return &pullRequestController{
		pullRequestUseCase: pullRequestUseCase,
	}
}


func (s *pullRequestController) PostPullRequestCreate(c *gin.Context) {
	var req gen.PostPullRequestCreateJSONBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pr, err := s.pullRequestUseCase.CreatePR(c.Request.Context(), domain.UserID(req.AuthorId),
					domain.PRID(req.PullRequestId), req.PullRequestName)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	prDTO := mapper.DomainPullRequestToDTO(*pr)

	c.JSON(http.StatusCreated, gin.H{"pr": prDTO})
}

func (s *pullRequestController) PostPullRequestMerge(c *gin.Context) {
	var req gen.PostPullRequestMergeJSONBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := s.pullRequestUseCase.MergePR(c.Request.Context(), domain.PRID(req.PullRequestId))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pr": req.PullRequestId})
}

func (s *pullRequestController) PostPullRequestReassign(c *gin.Context) {
	var req gen.PostPullRequestReassignJSONBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := s.pullRequestUseCase.ReassignReviewer(c.Request.Context(), domain.PRID(req.PullRequestId), domain.UserID(req.OldUserId))
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func (s *pullRequestController) GetUsersGetReview(c *gin.Context, params gen.GetUsersGetReviewParams) {
	prs, err := s.pullRequestUseCase.GetPRByUserID(c.Request.Context(), domain.UserID(params.UserId))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":       params.UserId,
		"pull_requests": mapper.DomainPullRequestsToDTOs(prs),
	})
}
