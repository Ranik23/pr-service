package integration_test

import (
	"app/internal/domain"
	"context"
	"database/sql"
)

func (s *TestSuite) Test_CreateTeam_Integration() {
	teamName := "integration-team"
	users := []domain.TeamUser{
		{ID: "user1", Name: "User One"},
		{ID: "user2", Name: "User Two"},
	}

	team, err := s.teamUseCase.CreateTeam(context.TODO(), teamName, users)
	s.Require().NoError(err)
	s.Require().NotNil(team)

	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)
	defer func() {
        if err = db.Close(); err != nil {
            s.T().Fatalf("failed to close db: %v", err)
        }   
    }()

	var dbTeamID int64
	var dbTeamName string
	err = db.QueryRow(`
		SELECT id, team_name FROM teams WHERE team_name = $1
	`, teamName).Scan(&dbTeamID, &dbTeamName)
	s.Require().NoError(err)
	s.Require().Equal(teamName, dbTeamName)

	for _, user := range users {
		var userCount int
		err = db.QueryRow(`
			SELECT COUNT(*) FROM users WHERE id = $1
		`, user.ID).Scan(&userCount)
		s.Require().NoError(err)
		s.Require().Equal(1, userCount)
	}

	for _, user := range users {
		var linkCount int
		err = db.QueryRow(`
			SELECT COUNT(*) FROM user_teams WHERE team_id = $1 AND user_id = $2
		`, dbTeamID, user.ID).Scan(&linkCount)
		s.Require().NoError(err)
		s.Require().Equal(1, linkCount)
	}
}

func (s *TestSuite) Test_CreateTeam_DuplicateTeamName_Integration() {
	teamName := "duplicate-team-integration"
	users := []domain.TeamUser{{ID: "user1", Name: "User One"}}

	team1, err := s.teamUseCase.CreateTeam(context.TODO(), teamName, users)
	s.Require().NoError(err)
	s.Require().NotNil(team1)

	team2, err := s.teamUseCase.CreateTeam(context.TODO(), teamName, []domain.TeamUser{{ID: "user2"}})
	s.Require().Error(err)
	s.Require().Nil(team2)

	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)
	defer func() {
        if err = db.Close(); err != nil {
            s.T().Fatalf("failed to close db: %v", err)
        }   
    }()

	var teamCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM teams WHERE team_name = $1
	`, teamName).Scan(&teamCount)
	s.Require().NoError(err)
	s.Require().Equal(1, teamCount)
}

func (s *TestSuite) Test_CreateTeam_UserAlreadyInTeam_Integration() {

	team1Name := "team1-integration"
	sharedUser := domain.UserID("shared-user-integration")
	team1, err := s.teamUseCase.CreateTeam(context.TODO(), team1Name, []domain.TeamUser{{ID: sharedUser, Name: "Shared User"}})
	s.Require().NoError(err)
	s.Require().NotNil(team1)

	team2Name := "team2-integration"
	team2, err := s.teamUseCase.CreateTeam(context.TODO(), team2Name, []domain.TeamUser{{ID: sharedUser, Name: "Shared User"}})
	s.Require().Error(err)
	s.Require().Nil(team2)

	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)
	defer func() {
        if err = db.Close(); err != nil {
            s.T().Fatalf("failed to close db: %v", err)
        }   
    }()

	var teamID int64
	err = db.QueryRow(`
		SELECT team_id FROM user_teams WHERE user_id = $1
	`, sharedUser).Scan(&teamID)
	s.Require().NoError(err)
	s.Require().Equal(team1.ID.Int64(), teamID)
}

func (s *TestSuite) Test_TeamWorkflow_Integration() {

	teamName := "workflow-team"
	users := []domain.TeamUser{
		{ID: "workflow-user1", Name: "Workflow User One"},
		{ID: "workflow-user2", Name: "Workflow User Two"},
		{ID: "workflow-user3", Name: "Workflow User Three"},
	}

	team, err := s.teamUseCase.CreateTeam(context.TODO(), teamName, users)
	s.Require().NoError(err)
	s.Require().NotNil(team)

	retrievedTeam, err := s.teamUseCase.GetTeamByName(context.TODO(), teamName)
	s.Require().NoError(err)
	s.Require().NotNil(retrievedTeam)

	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)
	defer func() {
        if err = db.Close(); err != nil {
            s.T().Fatalf("failed to close db: %v", err)
        }   
    }()

	s.Require().Equal(team.ID, retrievedTeam.ID)
	s.Require().Equal(team.TeamName, retrievedTeam.TeamName)
	s.Require().Len(retrievedTeam.Users, len(users))

	var dbUserCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM user_teams WHERE team_id = $1
	`, team.ID).Scan(&dbUserCount)
	s.Require().NoError(err)
	s.Require().Equal(len(users), dbUserCount)
}
