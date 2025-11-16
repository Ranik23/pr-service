package integration_test

import (
	"app/internal/domain"
	"context"
	"database/sql"
)

func (s *TestSuite) Test_ReassignReviewer_Integration_Success() {

    teamName := "reassign-team"
    authorID := domain.UserID("reassign-author")
    oldReviewer := domain.UserID("old-reviewer")
    extraUser := domain.UserID("extra-user")
    
    _, err := s.teamUseCase.CreateTeam(context.TODO(), teamName, []domain.TeamUser{
        {ID: authorID, Name: "Reassign Author"},
        {ID: oldReviewer, Name: "Old Reviewer"},
        {ID: extraUser, Name: "Extra User"},
    })
    s.Require().NoError(err)

    prName := "reassign-pr"
    pr, err := s.prUseCase.CreatePR(context.TODO(), authorID, "1", prName)
    s.Require().NoError(err)
    s.Require().NotNil(pr)

    db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
    s.Require().NoError(err)
    defer func() {
        if err = db.Close(); err != nil {
            s.T().Fatalf("failed to close db: %v", err)
        }   
    }()

    reviewerIDToReplace := oldReviewer

    err = s.prUseCase.ReassignReviewer(context.TODO(), pr.ID, domain.UserID(reviewerIDToReplace))
    s.Require().NoError(err)

    var oldReviewerCount int
    err = db.QueryRow(`
        SELECT COUNT(*) 
        FROM pr_reviewers 
        WHERE pr_id = $1 AND reviewer_id = $2
    `, pr.ID, reviewerIDToReplace.String()).Scan(&oldReviewerCount)
    s.Require().NoError(err)
    s.Require().Equal(0, oldReviewerCount)
}

func (s *TestSuite) Test_ReassignReviewer_Integration_ReviewerNotFound() {
    // Подготовка команды и PR
    teamName := "reassign-notfound-team"
    authorID := domain.UserID("author-notfound")
    _, err := s.teamUseCase.CreateTeam(context.TODO(), teamName, []domain.TeamUser{
        {ID: authorID, Name: "Author NotFound"},
        {ID: "reviewer-notfound", Name: "Reviewer NotFound"},
    })
    
    s.Require().NoError(err)

    prName := "reassign-notfound-pr"
    pr, err := s.prUseCase.CreatePR(context.TODO(), authorID, "1", prName)
    s.Require().NoError(err)
    s.Require().NotNil(pr)

    // Пытаемся переприсвоить несуществующего ревьювера
    err = s.prUseCase.ReassignReviewer(context.TODO(), pr.ID, "non-existent-reviewer")
    s.Require().Error(err)
}