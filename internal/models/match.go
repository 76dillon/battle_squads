package models

import "time"

type MatchState string

const (
	MatchStatePending    MatchState = "PENDING"
	MatchStateInProgress            = "IN_PROGRESS"
	MatchStateCompleted             = "COMPLETED"
)

type Match struct {
	ID                   int64
	State                MatchState
	Player1ID            int64
	Player2ID            int64
	WinnerPlayerID       *int64
	CurrentTurnNumber    int
	CurrentActorPlayerID *int64
	CreatedAt            time.Time
	StartedAt            *time.Time
	CompletedAt          *time.Time
}

type MatchSide struct {
	ID          int64
	MatchID     int64
	PlayerID    int64
	SquadID     int64
	ActiveIndex int
}

type MatchUnit struct {
	ID          int64
	MatchSideID int64
	UnitID      int64
	Position    int
	CurrentHP   int
}
