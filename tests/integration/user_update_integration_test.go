package integration_test

import (
	"app/internal/domain"
	"context"
	"database/sql"
)

func (s *TestSuite) Test_UpdateUserActivity_Success_Activate() {
    userID := domain.UserID("test-update-activate")
    name := "Test User"
    
    _, err := s.userUseCase.CreateUser(context.TODO(), userID, name)
    s.Require().NoError(err)

    _, err = s.userUseCase.UpdateUserActivity(context.TODO(), userID, domain.UserStatusInactive)
    s.Require().NoError(err)

    db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
    s.Require().NoError(err)
    defer func() {
        if err = db.Close(); err != nil {
            s.T().Fatalf("failed to close db: %v", err)
        }   
    }()

    var isActive bool
    err = db.QueryRow(`
        SELECT is_active 
        FROM users 
        WHERE id = $1
    `, userID).Scan(&isActive)
    s.Require().NoError(err)
    s.Require().False(isActive)

    _, err = s.userUseCase.UpdateUserActivity(context.TODO(), userID, domain.UserStatusActive)
    s.Require().NoError(err)

    err = db.QueryRow(`
        SELECT is_active 
        FROM users 
        WHERE id = $1
    `, userID).Scan(&isActive)
    s.Require().NoError(err)
    s.Require().True(isActive)
}
