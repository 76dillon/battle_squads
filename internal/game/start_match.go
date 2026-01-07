package game

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"

	"github.com/76dillon/battle_squads/internal/store"
)

// StartMatch sets up match_sides and match_units based on chosen squads.
func (s *Service) StartMatch(
	ctx context.Context,
	matchID int64,
	p1SquadID int64,
	p2SquadID int64,
) error {
	// 1. Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	qtx := s.q.WithTx(tx)
	// use qtx everywhere instead of s.q
	// commit or rollback

	match, err := qtx.GetMatchByID(ctx, matchID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("error retrieving match information: %w", err)
	}

	// 2. For player 1:
	//    - CreateMatchSide(matchID, p1ID, p1SquadID, active_index = 0)
	p1MatchSide, err := qtx.CreateMatchSide(ctx, store.CreateMatchSideParams{
		MatchID:  match.ID,
		PlayerID: match.Player1ID,
		SquadID:  p1SquadID,
	})
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("error creating match side: %w", err)
	}
	//    - GetSquadUnits(p1SquadID)
	p1SquadUnits, err := qtx.GetSquadUnits(ctx, p1SquadID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("error retrieving squad units: %w", err)
	}

	//    - For each squad unit:
	for _, squadUnit := range p1SquadUnits {
		//        - GetUnitByID(unit_id)
		unit, err := qtx.GetUnitByID(ctx, squadUnit.UnitID)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("error retrieving unit info: %w", err)
		}
		//        - CreateMatchUnit(side1.ID, unit_id, position, current_hp = base_hp)
		_, err = qtx.CreateMatchUnit(ctx, store.CreateMatchUnitParams{
			MatchSideID: p1MatchSide.ID,
			UnitID:      unit.ID,
			Position:    squadUnit.Position,
			CurrentHp:   unit.BaseHp,
		})
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("error creating match unit: %w", err)
		}
	}

	// 3. For player 2: same pattern as player 1
	//    - CreateMatchSide(matchID, p2ID, p2SquadID, active_index = 0)
	p2MatchSide, err := qtx.CreateMatchSide(ctx, store.CreateMatchSideParams{
		MatchID:  match.ID,
		PlayerID: match.Player2ID,
		SquadID:  p2SquadID,
	})
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("error creating match side: %w", err)
	}
	//    - GetSquadUnits(p2SquadID)
	p2SquadUnits, err := qtx.GetSquadUnits(ctx, p2SquadID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("error retrieving squad units: %w", err)
	}

	//    - For each squad unit:
	for _, squadUnit := range p2SquadUnits {
		//        - GetUnitByID(unit_id)
		unit, err := qtx.GetUnitByID(ctx, squadUnit.UnitID)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("error retrieving unit info: %w", err)
		}
		//        - CreateMatchUnit(side1.ID, unit_id, position, current_hp = base_hp)
		_, err = qtx.CreateMatchUnit(ctx, store.CreateMatchUnitParams{
			MatchSideID: p2MatchSide.ID,
			UnitID:      unit.ID,
			Position:    squadUnit.Position,
			CurrentHp:   unit.BaseHp,
		})
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("error creating match unit: %w", err)
		}
	}

	// 4. Determine initial active units:
	//    - find match_units where side = side1 and position = 0
	p1Active, err := qtx.GetActiveMatchUnitForSide(ctx, p1MatchSide.ID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("get p1 active unit: %w", err)
	}

	//    - find match_units where side = side2 and position = 0
	p2Active, err := qtx.GetActiveMatchUnitForSide(ctx, p2MatchSide.ID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("get p2 active unit: %w", err)
	}

	//    - load corresponding units to get base_speed
	p1Unit, err := qtx.GetUnitByID(ctx, p1Active.UnitID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("get p1 unit: %w", err)
	}
	p2Unit, err := qtx.GetUnitByID(ctx, p2Active.UnitID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("get p2 unit: %w", err)
	}
	p1Speed := p1Unit.BaseSpeed
	p2Speed := p2Unit.BaseSpeed
	// 5. Compare speeds, random tie-breaker:
	//    - decide initialActorPlayerID
	var initialActor int64
	switch {
	case p1Speed > p2Speed:
		initialActor = match.Player1ID
	case p2Speed > p1Speed:
		initialActor = match.Player2ID
	default:
		if rand.Intn(2) == 0 {
			initialActor = match.Player1ID
		} else {
			initialActor = match.Player2ID
		}
	}
	// 6. Call StartMatch (UPDATE matches SET state='IN_PROGRESS', ...)
	_, err = qtx.StartMatch(ctx, store.StartMatchParams{
		ID: match.ID,
		CurrentActorPlayerID: sql.NullInt64{
			Int64: initialActor,
			Valid: true,
		},
	})
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("update match to in_progress: %w", err)
	}

	// 7. Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
