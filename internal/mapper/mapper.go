package mapper

import (
	"time"

	"app/internal/controllers/gen"
	"app/internal/domain"
	"app/internal/repository/models"
)

func DomainUserToModel(u domain.User) models.User {
    return models.User{
        ID:             u.ID,
        Name:           u.Name,
        StatusActivity: u.IsActive.IsActive(),
    }
}

func DomainUsersToModels(users []domain.User) []models.User {
	result := make([]models.User, 0, len(users))
	for _, user := range users {
		result = append(result, DomainUserToModel(user))
	}
	return result
}

func DomainToModelTeam(team domain.Team) models.Team {
	return models.Team{
		ID:        team.ID,
		TeamName:  team.TeamName,
		CreatedAt: team.CreatedAt,
	}
}

func DomainToModelTeams(teams []domain.Team) []models.Team {
	result := make([]models.Team, 0, len(teams))
	for _, team := range teams {
		result = append(result, DomainToModelTeam(team))
	}
	return result
}

func DomainToModelPullRequest(pr domain.PullRequest) models.PullRequest {
	return models.PullRequest{
		ID:                pr.ID,
		Name:              pr.Name,
		AuthorID:          pr.Author.ID,
		Status:            pr.Status,
		NeedMoreReviewers: pr.NeedMoreReviewers,
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}

func DomainToModelPullRequests(prs []domain.PullRequest) []models.PullRequest {
	result := make([]models.PullRequest, 0, len(prs))
	for _, pr := range prs {
		result = append(result, DomainToModelPullRequest(pr))
	}
	return result
}


func ModelToDomainUser(user models.User) domain.User {
	status := domain.UserStatusInactive
	if user.StatusActivity {
		status = domain.UserStatusActive
	}

	return domain.User{
		ID:       user.ID,
		Name:     user.Name,
		IsActive: status,
	}
}

func ModelsToDomainUsers(users []models.User) []domain.User {
	result := make([]domain.User, 0, len(users))
	for _, user := range users {
		result = append(result, ModelToDomainUser(user))
	}
	return result
}

func ModelToDomainTeam(team models.Team, teamUsers []models.User) domain.Team {
	var domainUsers []domain.User
	for _, user := range teamUsers {
		domainUsers = append(domainUsers, ModelToDomainUser(user))
	}

	return domain.Team{
		ID:       team.ID,
		TeamName: team.TeamName,
		Users:    domainUsers,
	}
}

func ModelsToDomainTeams(teams []models.Team, teamUsersMap map[domain.TeamID][]models.User) []domain.Team {
	result := make([]domain.Team, 0, len(teams))
	for _, team := range teams {
		users := teamUsersMap[team.ID]
		result = append(result, ModelToDomainTeam(team, users))
	}
	return result

}

func ModelToDomainPullRequest(pr models.PullRequest, author models.User, reviewers []models.User) domain.PullRequest {
	var domainReviewers []domain.User
	for _, reviewer := range reviewers {
		domainReviewers = append(domainReviewers, ModelToDomainUser(reviewer))
	}

	return domain.PullRequest{
		ID:                pr.ID,
		Name:              pr.Name,
		Author:            ModelToDomainUser(author),
		Status:            pr.Status,
		Reviewers:         domainReviewers,
		NeedMoreReviewers: pr.NeedMoreReviewers,
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}


func ModelsToDomainPullRequests(prs []models.PullRequest, authorsMap map[domain.UserID]models.User, reviewersMap map[domain.PRID][]models.User) []domain.PullRequest {
	result := make([]domain.PullRequest, 0, len(prs))
	for _, pr := range prs {
		author := authorsMap[pr.AuthorID]
		reviewers := reviewersMap[pr.ID]
		result = append(result, ModelToDomainPullRequest(pr, author, reviewers))
	}
	return result
}


func DomainUserToDTO(user domain.User, teamName string) gen.User {
	return gen.User{
		UserId:   user.ID.String(),
		Username: user.Name,
		IsActive: user.IsActive.IsActive(),
		TeamName: teamName,
	}
}

func DomainUsersToDTOs(users []domain.User, teamName string) []gen.User {
	result := make([]gen.User, 0, len(users))
	for _, user := range users {
		result = append(result, DomainUserToDTO(user, teamName))
	}
	return result
}


func DomainTeamToDTO(team domain.Team) gen.Team {
	return gen.Team{
		TeamName: team.TeamName,
		Members:  DomainTeamMembersToDTO(team.Users),
	}
}

func DomainTeamMembersToDTO(users []domain.User) []gen.TeamMember {
	result := make([]gen.TeamMember, 0, len(users))
	for _, user := range users {
		result = append(result, gen.TeamMember{
			UserId:   user.ID.String(),
			Username: user.Name,
			IsActive: user.IsActive.IsActive(),
		})
	}
	return result
}


func DomainPullRequestToDTO(pr domain.PullRequest) gen.PullRequest {
	var createdAt, mergedAt *time.Time
	
	if !pr.CreatedAt.IsZero() {
		createdAt = &pr.CreatedAt
	}
	if pr.MergedAt != nil && !pr.MergedAt.IsZero() {
		mergedAt = pr.MergedAt
	}

	assignedReviewers := make([]string, 0, len(pr.Reviewers))
	for _, reviewer := range pr.Reviewers {
		assignedReviewers = append(assignedReviewers, reviewer.ID.String())
	}

	return gen.PullRequest{
		PullRequestId:   pr.ID.String(),
		PullRequestName: pr.Name,
		AuthorId:        pr.Author.ID.String(),
		Status:          gen.PullRequestStatus(pr.Status.String()),
		AssignedReviewers: assignedReviewers,
		CreatedAt:       createdAt,
		MergedAt:        mergedAt,
	}
}

func DomainPullRequestsToDTOs(prs []domain.PullRequest) []gen.PullRequest {
	result := make([]gen.PullRequest, 0, len(prs))
	for _, pr := range prs {
		result = append(result, DomainPullRequestToDTO(pr))
	}
	return result
}

func DomainUserStatsToDTO(stats domain.UserStats) gen.UserAssignmentStats {
	return gen.UserAssignmentStats{
		UserId:        stats.UserID.String(),
		AssignedCount: stats.AssignedCount,
	}
}


func DomainUserStatsToDTOs(stats []domain.UserStats) []gen.UserAssignmentStats {
	result := make([]gen.UserAssignmentStats, 0, len(stats))
	for _, stat := range stats {
		result = append(result, DomainUserStatsToDTO(stat))
	}
	return result
}

