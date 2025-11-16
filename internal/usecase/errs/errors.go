package errs

import "errors"

var (
	ErrPRAlreadyMerged 					= errors.New("pull request already merged")
	ErrInvalidPullRequestName		    = errors.New("invalid pull request name")
	ErrInvalidPullRequestID             = errors.New("invalid pull request ID")
	ErrInvalidTeamName 				    = errors.New("invalid team name")
	ErrNoUsersProvided 					= errors.New("no users provided for team creation")
	ErrNoAvailableActiveUserToAssign    = errors.New("no available active user to assign")
	ErrNoUsersInTeam 				    = errors.New("no users in team")
	ErrUserAlreadyHasTeam 				= errors.New("one of the users already has a team")
	ErrAuthorNotFound     				= errors.New("author not found")
	ErrUserNotFound    					= errors.New("user not found")
	ErrTeamNotFound    					= errors.New("team not found")
	ErrUserAlreadyExists      			= errors.New("user already exists")
	ErrTeamAlreadyExists      			= errors.New("team already exists")
	ErrInvalidInput    					= errors.New("invalid input provided")
	ErrPullRequestNotFound    			= errors.New("pull request not found")
	ErrPullRequestAlreadyExists 		= errors.New("pull request already exists")
	ErrPullRequestAlreadyMerged 		= errors.New("pull request already merged")
	ErrUserHasNoTeam 					= errors.New("user has no team")
	ErrReviewerNotFoundInPR 			= errors.New("reviewer not found in pull request")
	ErrReviewerNotFoundInPullRequest 	= errors.New("reviewer not found in pull request")
	ErrInvalidUserID					= errors.New("invalid user id")
)