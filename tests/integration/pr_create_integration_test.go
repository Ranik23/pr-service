package integration_test

import (
	"app/internal/domain"
	"app/internal/usecase/errs"
	"context"
	"database/sql"
	"errors"
	"strings"
)

func (s *TestSuite) Test_CreatePR_Integration_TransactionRollback() {
    // Подготовка команды
    teamName := "rollback-pr-team"
    authorID := domain.UserID("author-rollback")
    invalidUser := domain.UserID(strings.Repeat("a", 300))
    
    _, err := s.teamUseCase.CreateTeam(context.TODO(), teamName, []domain.TeamUser{
        {ID: authorID, Name: "Author Rollback"},
        {ID: invalidUser, Name: "Invalid User"},
    })

	s.Require().True(errors.Is(err, errs.ErrInvalidUserID))

    prName := "rollback-pr"
    pr, err := s.prUseCase.CreatePR(context.TODO(), authorID, "1", prName)
    s.Require().Error(err)
    s.Require().Nil(pr)

    db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
    s.Require().NoError(err)
    defer func() {
        if err := db.Close(); err != nil {
            s.T().Logf("Failed to close database connection: %v", err)
        }
    }()

    var count int
    err = db.QueryRow(`
        SELECT COUNT(*) 
        FROM pull_requests 
        WHERE name = $1
    `, prName).Scan(&count)
    s.Require().NoError(err)
    s.Require().Equal(0, count)
}

func (s *TestSuite) Test_CreatePR_Integration_Success() {
    teamName := "pr-team"
    authorID := domain.UserID("author-user")
    reviewer1 := domain.UserID("reviewer-1")
    reviewer2 := domain.UserID("reviewer-2")
    
    _, err := s.teamUseCase.CreateTeam(context.TODO(), teamName, []domain.TeamUser{{
        ID: authorID, 
        Name: "Author User",
    }, {
        ID: reviewer1, 
        Name: "Reviewer 1",
    }, {
        ID: reviewer2, 
        Name: "Reviewer 2",
    }})

    s.Require().NoError(err)

    prName := "test-pr"
    pr, err := s.prUseCase.CreatePR(context.TODO(), authorID, "1", prName)
    s.Require().NoError(err)
    s.Require().NotNil(pr)

    db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
    s.Require().NoError(err)
    defer func() {
        if err := db.Close(); err != nil {
            s.T().Logf("Failed to close database connection: %v", err)
        }
    }()

    var dbPRID int64
    var dbPRName, dbAuthorID string
    var dbStatus string
    err = db.QueryRow(`
        SELECT id, name, author_id, status 
        FROM pull_requests 
        WHERE name = $1
    `, prName).Scan(&dbPRID, &dbPRName, &dbAuthorID, &dbStatus)
    s.Require().NoError(err)
    s.Require().Equal(prName, dbPRName)
    s.Require().Equal(authorID.String(), dbAuthorID)
    s.Require().Equal("OPEN", dbStatus)

	// Проверяем что назначены ревьюверы от 1 до 2
    var reviewerCount int
    err = db.QueryRow(`
        SELECT COUNT(*) 
        FROM pr_reviewers 
        WHERE pr_id = $1
    `, dbPRID).Scan(&reviewerCount)
    s.Require().NoError(err)
    s.Require().GreaterOrEqual(reviewerCount, 1)
    s.Require().LessOrEqual(reviewerCount, 2)

    var authorAsReviewer int
    err = db.QueryRow(`
        SELECT COUNT(*) 
        FROM pr_reviewers 
        WHERE pr_id = $1 AND reviewer_id = $2
    `, dbPRID, authorID).Scan(&authorAsReviewer)
    s.Require().NoError(err)
    s.Require().Equal(0, authorAsReviewer)
}

func (s *TestSuite) Test_CreatePR_Integration_UserNotInTeam() {
    userID := domain.UserID("lonely-user")
    name   := "Lonely User"

    _, err := s.userUseCase.CreateUser(context.TODO(), userID, name)
    s.Require().NoError(err)

    prName := "lonely-pr"
    pr, err := s.prUseCase.CreatePR(context.TODO(), userID, "1", prName)
    s.Require().Error(err)
    s.Require().Nil(pr)

    db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
    s.Require().NoError(err)
    defer func() {
        if err := db.Close(); err != nil {
            s.T().Logf("Failed to close database connection: %v", err)
        }
    }()

    var count int
    err = db.QueryRow(`
        SELECT COUNT(*) 
        FROM pull_requests 
        WHERE name = $1
    `, prName).Scan(&count)
    s.Require().NoError(err)
    s.Require().Equal(0, count)
}

func (s *TestSuite) Test_CreatePR_Integration_DuplicatePRName() {
    teamName := "duplicate-pr-team"
    authorID := domain.UserID("author-dup")
    _, err := s.teamUseCase.CreateTeam(context.TODO(), teamName, []domain.TeamUser{{
        ID: authorID, 
        Name: "Author Dup",
    }, {
        ID: "reviewer-dup", 
        Name: "Reviewer Dup",
    }})
    
    s.Require().NoError(err)

    prName := "duplicate-pr"
    pr1, err := s.prUseCase.CreatePR(context.TODO(), authorID, "1", prName)
    s.Require().NoError(err)
    s.Require().NotNil(pr1)

    pr2, err := s.prUseCase.CreatePR(context.TODO(), authorID, "2", prName)
    s.Require().Error(err)
    s.Require().Nil(pr2)

    db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
    s.Require().NoError(err)
    defer func() {
        if err := db.Close(); err != nil {
        s.T().Logf("Failed to close database connection: %v", err)
        }
    }()

    var count int
    err = db.QueryRow(`
        SELECT COUNT(*) 
        FROM pull_requests 
        WHERE name = $1
    `, prName).Scan(&count)
    s.Require().NoError(err)
    s.Require().Equal(1, count)
}

