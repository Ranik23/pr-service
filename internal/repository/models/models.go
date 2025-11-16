package models

import (
	"app/internal/domain"
	"time"
)

type User struct {
	ID        		domain.UserID
	Name      		string     
	StatusActivity  bool      
}

type Team struct {
	ID        domain.TeamID
	TeamName  string    
	CreatedAt time.Time
}

type UserTeam struct {
	UserID 		domain.UserID
	TeamID 		domain.TeamID
	CreatedAt 	time.Time
}

type PullRequest struct {
	ID                domain.PRID   
	Name              string    
	AuthorID          domain.UserID    
	Status            domain.PRStatus  
	NeedMoreReviewers bool     
	CreatedAt         time.Time 
	MergedAt          *time.Time 
}


type PullRequestReviewer struct {
	PullRequestID domain.PRID
	ReviewerID    domain.UserID
	CreatedAt     time.Time
}