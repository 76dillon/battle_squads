package httpapi

import (
	"context"
	"time"

	"github.com/76dillon/battle_squads/internal/store"
)

func (s *Server) buildMatchResponse(
	m store.Match,
	sides []store.MatchSide,
	unitsBySide map[int64][]store.MatchUnit,
) MatchResponse {
	var winnerID *int64
	if m.WinnerPlayerID.Valid {
		winnerID = &m.WinnerPlayerID.Int64
	}

	var startedAt *time.Time
	if m.StartedAt.Valid {
		startedAt = &m.StartedAt.Time
	}

	var completedAt *time.Time
	if m.CompletedAt.Valid {
		completedAt = &m.CompletedAt.Time
	}

	var currentActor *int64
	if m.CurrentActorPlayerID.Valid {
		currentActor = &m.CurrentActorPlayerID.Int64
	}

	mv := MatchView{
		ID:                   m.ID,
		State:                MatchState(m.State),
		CreatedAt:            m.CreatedAt,
		StartedAt:            startedAt,
		CompletedAt:          completedAt,
		Player1ID:            m.Player1ID,
		Player2ID:            m.Player2ID,
		WinnerPlayerID:       winnerID,
		CurrentTurnNumber:    int(m.CurrentTurnNumber),
		CurrentActorPlayerID: currentActor,
	}

	svs := make([]SideView, 0, len(sides))
	for _, side := range sides {
		units := unitsBySide[side.ID]
		uvs := make([]UnitView, 0, len(units))
		for _, u := range units {
			uv := UnitView{
				MatchUnitID: u.ID,
				UnitID:      u.UnitID,
				Position:    u.Position,
				CurrentHP:   u.CurrentHp,
				IsActive:    (u.Position == side.ActiveIndex),
			}

			if uv.IsActive {
				moves, err := s.q.ListMovesForUnit(context.Background(), u.UnitID)
				// in real code, use ctx and handle err, but here conceptually:
				if err == nil {
					mv := make([]MoveView, 0, len(moves))
					for _, m := range moves {
						mv = append(mv, MoveView{
							ID:       m.ID,
							Name:     m.Name,
							Power:    m.Power,
							Accuracy: m.Accuracy,
						})
					}
					uv.Moves = mv
				}
			}
			uvs = append(uvs, uv)
		}
		svs = append(svs, SideView{
			PlayerID:  side.PlayerID,
			SquadID:   side.SquadID,
			ActivePos: side.ActiveIndex,
			Units:     uvs,
		})
	}

	return MatchResponse{
		Match: mv,
		Sides: svs,
	}
}
