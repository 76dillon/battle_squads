package httpapi

import "time"

type MatchState string

type MatchView struct {
	ID                   int64      `json:"id"`
	State                MatchState `json:"state"`
	CreatedAt            time.Time  `json:"created_at"`
	StartedAt            *time.Time `json:"started_at,omitempty"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
	Player1ID            int64      `json:"player1_id"`
	Player2ID            int64      `json:"player2_id"`
	WinnerPlayerID       *int64     `json:"winner_player_id,omitempty"`
	CurrentTurnNumber    int        `json:"current_turn_number"`
	CurrentActorPlayerID *int64     `json:"current_actor_player_id,omitempty"`
}

type UnitView struct {
	MatchUnitID int64      `json:"match_unit_id"`
	UnitID      int64      `json:"unit_id"`
	Position    int32      `json:"position"`
	CurrentHP   int32      `json:"current_hp"`
	IsActive    bool       `json:"is_active"`
	Moves       []MoveView `json:"moves,omitempty"`
}

type SideView struct {
	PlayerID  int64      `json:"player_id"`
	SquadID   int64      `json:"squad_id"`
	ActivePos int32      `json:"active_position"`
	Units     []UnitView `json:"units"`
}

type MatchResponse struct {
	Match MatchView  `json:"match"`
	Sides []SideView `json:"sides"`
}

type MoveView struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Power    int32  `json:"power"`
	Accuracy int32  `json:"accuracy"`
}
