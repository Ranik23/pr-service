package domain

type PRStatus string

func (s PRStatus) String() string {
    return string(s)
}

const (
    PRStatusOpen   PRStatus = "OPEN"
    PRStatusMerged PRStatus = "MERGED"
)

type PRID string

func (id PRID) String() string {
    return string(id)
}

type UserID string

func (id UserID) String() string {
    return string(id)
}

type UserActivityStatus bool

const (
    UserStatusActive   UserActivityStatus = true
    UserStatusInactive UserActivityStatus = false
)

func (s UserActivityStatus) IsActive() bool {
    return s == UserStatusActive
}

func (s UserActivityStatus) String() string {
    if s == UserStatusActive {
        return "active"
    }
    return "inactive"
}



type TeamID int64

func (id TeamID) Int64() int64 {
    return int64(id)
}
