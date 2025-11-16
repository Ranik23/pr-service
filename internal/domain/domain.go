package domain

import (
	"time"
)

type User struct {
	ID       UserID
	IsActive UserActivityStatus
	Name 	 string
}

type TeamUser struct {
	ID   UserID
	Name string
}

type Team struct {
	ID        TeamID
	TeamName  string
	Users     []User
	CreatedAt time.Time
}

type PullRequest struct {
	ID                PRID
	Name              string
	Author            User
	Status            PRStatus
	Reviewers         []User
	NeedMoreReviewers bool
	CreatedAt         time.Time
	MergedAt          *time.Time
}

type UserStats struct {
	UserID        UserID
	AssignedCount int
}
