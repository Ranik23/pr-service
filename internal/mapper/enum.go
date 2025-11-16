package mapper

import (
	"app/internal/controllers/gen"
	"app/internal/domain"
)

func FromDomainPullRequestShortStatusToDTO(status domain.PRStatus) gen.PullRequestShortStatus {
	switch status {
	case domain.PRStatusOpen:
		return gen.PullRequestShortStatusOPEN
	case domain.PRStatusMerged:
		return gen.PullRequestShortStatusMERGED
	default:
		return gen.PullRequestShortStatusOPEN
	}
}

func FromDomainPullRequestStatusToDTO(status domain.PRStatus) gen.PullRequestStatus {
	switch status {
	case domain.PRStatusOpen:
		return gen.PullRequestStatusOPEN
	case domain.PRStatusMerged:
		return gen.PullRequestStatusMERGED
	default:
		return gen.PullRequestStatusOPEN
	}
}
