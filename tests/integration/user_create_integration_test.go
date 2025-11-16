package integration_test

import (
	"app/internal/domain"
	"app/internal/repository/models"
	"app/internal/usecase/errs"
	"context"
	"database/sql"
	"errors"
	"strings"
)

func (s *TestSuite) Test_CreateUser_Success() {
	userID := domain.UserID("test-user-id")
	name := "Test User"

	_, err := s.userUseCase.CreateUser(context.TODO(), userID, name)
	s.Require().NoError(err)

	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)
	defer func() {
        if err = db.Close(); err != nil {
            s.T().Fatalf("failed to close db: %v", err)
        }   
    }()

	var userModel models.User
	err = db.QueryRow(`
    	SELECT id, name, is_active
    	FROM users
    	WHERE id = $1
	`, userID).Scan(&userModel.ID, &userModel.Name, &userModel.StatusActivity)
	s.Require().NoError(err)

	s.Require().Equal(userID, userModel.ID)
	s.Require().Equal(name, userModel.Name)
	s.Require().True(userModel.StatusActivity)
}


func (s *TestSuite) Test_CreateUser_DuplicateUserID() {
	userID := domain.UserID("duplicate-user-id")
	name := "Duplicate User"
	

	user1, err := s.userUseCase.CreateUser(context.TODO(), userID, name)
	s.Require().NoError(err)
	s.Require().NotNil(user1)

	user2, err := s.userUseCase.CreateUser(context.TODO(), userID, name)
	s.Require().Error(err)
	s.Require().Nil(user2)
	s.Require().True(errors.Is(err, errs.ErrUserAlreadyExists))

	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)
	defer func() {
        if err = db.Close(); err != nil {
            s.T().Fatalf("failed to close db: %v", err)
        }   
    }()

	var count int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM users 
		WHERE id = $1
	`, userID).Scan(&count)
	s.Require().NoError(err)
	s.Require().Equal(1, count)
}


func (s *TestSuite) Test_CreateUser_LongUserID() {
	userID := domain.UserID(strings.Repeat("a", 300))
	name := "Long User"
	

	_, err := s.userUseCase.CreateUser(context.TODO(), userID, name)
	s.Require().Error(err)
	s.Require().True(errors.Is(err, errs.ErrInvalidUserID))

	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)
	defer func() {
        if err = db.Close(); err != nil {
            s.T().Fatalf("failed to close db: %v", err)
        }   
    }()

	var dbUserID string
	err = db.QueryRow(`
		SELECT id
		FROM users
		WHERE id = $1
	`, userID).Scan(&dbUserID)

	s.Require().True(errors.Is(err, sql.ErrNoRows))
}