package integration_test

import (
	"app/internal/domain"
	"context"
	"database/sql"
)


func (s *TestSuite) Test_MergePR_Integration_Success() {
    // Подготовка команды и PR
    teamName := "merge-team"
    authorID := domain.UserID("merge-author")
    _, err := s.teamUseCase.CreateTeam(context.TODO(), teamName, []domain.TeamUser{{
        ID: authorID, 
        Name: "Merge Author",
    }, {
        ID: "reviewer-merge", 
        Name: "Reviewer Merge",
    }})
    s.Require().NoError(err)

    prName := "merge-pr"
    pr, err := s.prUseCase.CreatePR(context.TODO(), authorID, "1", prName)
    s.Require().NoError(err)
    s.Require().NotNil(pr)

    // Мержим PR
    err = s.prUseCase.MergePR(context.TODO(), pr.ID)
    s.Require().NoError(err)

    // Проверяем в БД что статус изменился на MERGED
    db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
    s.Require().NoError(err)
    defer func() {
        if err = db.Close(); err != nil {
            s.T().Fatalf("failed to close db: %v", err)
        }   
    }()

    var status string
    var mergedAt sql.NullTime
    err = db.QueryRow(`
        SELECT status, merged_at 
        FROM pull_requests 
        WHERE id = $1
    `, pr.ID).Scan(&status, &mergedAt)
    s.Require().NoError(err)
    s.Require().Equal("MERGED", status)
    s.Require().True(mergedAt.Valid)
}