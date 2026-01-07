package game

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"

	"github.com/76dillon/battle_squads/internal/store"
)

// ApplyTurn applies a player's action to the current match state.
func (s *Service) ApplyTurn(
	ctx context.Context,
	matchID int64,
	actingPlayerID int64,
	moveID int64,
) error {

	// 1. Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	qtx := s.q.WithTx(tx)

	//2. Load match by match ID
	match, err := qtx.GetMatchByID(ctx, matchID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("error retrieving match information: %w", err)
	}

	//3. Validate it's the right player's turn
	//--match.State must be in progress
	if match.State != "IN_PROGRESS" {
		_ = tx.Rollback()
		return ErrMatchNotInProgress{Msg: "match is not in progress"}
	}
	//--match.CurrentActorPlayerID must be equal to actingPlayerID
	if !match.CurrentActorPlayerID.Valid || match.CurrentActorPlayerID.Int64 != actingPlayerID {
		_ = tx.Rollback()
		return ErrWrongTurn{Msg: "it is not your turn yet"}
	}

	//4. Find acting player's match_side and active_unit
	//--GetMatchSidesByMatchID -> Pick one with PlayerID = actingPlayerID
	sides, err := qtx.GetMatchSidesByMatchID(ctx, match.ID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("error retrieving match sides: %w", err)
	}
	var actingSide *store.MatchSide
	for _, side := range sides {
		if side.PlayerID == actingPlayerID {
			actingSide = &side
			break
		}
	}
	if actingSide == nil {
		_ = tx.Rollback()
		return ErrIllegalMove{Msg: "no side found for acting player"}
	}

	//--GetActiveMatchUnitForSide(matchSide.ID) -> actingUnit
	actingMatchUnit, err := qtx.GetActiveMatchUnitForSide(ctx, actingSide.ID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("get acting active unit: %w", err)
	}

	//5. Validate the target
	//--find opponent side
	var opponentSide *store.MatchSide
	for i := range sides {
		if sides[i].PlayerID != actingPlayerID {
			opponentSide = &sides[i]
			break
		}
	}
	if opponentSide == nil {
		_ = tx.Rollback()
		return ErrIllegalMove{Msg: "no opponent side found"}
	}
	//--Load targetMatchUnit by GetActiveMatchUnitForSide(opponentSide.ID)
	targetMU, err := qtx.GetActiveMatchUnitForSide(ctx, opponentSide.ID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("error retrieving opponent active unit: %w", err)
	}

	//--Ensure targetMatchUnut.CurrentHp > 0
	if targetMU.CurrentHp <= 0 {
		_ = tx.Rollback()
		return ErrIllegalMove{Msg: "opponent's active unit is already KO'd"}
	}

	//6. Validate the move
	//--Ensure moveID is allowed for actingUnit.UnitID(ListMovesForUnit and check for a dedicated query)
	moves, err := qtx.ListMovesForUnit(ctx, actingMatchUnit.UnitID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("error retrieving moves for unit: %w", err)
	}

	var move *store.Move
	for i := range moves {
		if moves[i].ID == moveID {
			move = &moves[i]
			break
		}
	}

	if move == nil {
		_ = tx.Rollback()
		return ErrIllegalMove{Msg: "unit does not know this move"}
	}

	//7. Compute Damage
	//--Load move (power, type) and actingUnit's Unit (base_attack)
	attackerUnit, err := qtx.GetUnitByID(ctx, actingMatchUnit.UnitID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("get attacker unit: %w", err)
	}
	//--Damage rule: move.Power + unit.BaseAttack/2
	damage := move.Power + attackerUnit.BaseAttack/2
	if damage < 1 {
		damage = 1
	}

	//--Subtract damage from targetMatchUnit.CurrentHp, clamp at >=0
	newHP := targetMU.CurrentHp - damage
	if newHP < 0 {
		newHP = 0
	}
	//8. Update target HP in match_units
	updatedTargetMU, err := qtx.UpdateMatchUnitHP(ctx, store.UpdateMatchUnitHPParams{
		ID:        targetMU.ID,
		CurrentHp: newHP,
	})
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("update target hp: %w", err)
	}

	didKO := updatedTargetMU.CurrentHp == 0
	matchCompleted := false
	var winnerID int64
	//9. Check KO cases
	if didKO {
		// load units for the opponent side
		units, err := qtx.GetMatchUnitsBySideID(ctx, opponentSide.ID)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("get units for opponent side: %w", err)
		}

		nextIndex := int32(-1)
		for _, mu := range units {
			if mu.Position > opponentSide.ActiveIndex && mu.CurrentHp > 0 {
				nextIndex = mu.Position
				break
			}
		}

		if nextIndex == -1 {
			// no alive units left → mark match as completed and winner
			if opponentSide.PlayerID == match.Player1ID {
				winnerID = match.Player2ID
			} else {
				winnerID = match.Player1ID
			}
			matchCompleted = true
		} else {
			// auto-switch
			_, err := qtx.UpdateMatchSideActiveIndex(ctx, store.UpdateMatchSideActiveIndexParams{
				ID:          opponentSide.ID,
				ActiveIndex: nextIndex,
			})
			if err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("update active index: %w", err)
			}
		}
	}
	_, err = qtx.CreateMatchTurn(ctx, store.CreateMatchTurnParams{
		MatchID:           match.ID,
		TurnNumber:        match.CurrentTurnNumber,
		ActingPlayerID:    actingPlayerID,
		ActingMatchUnitID: actingMatchUnit.ID,
		MoveID:            moveID,
		TargetMatchUnitID: updatedTargetMU.ID,
		DamageDone:        damage,
		TargetHpAfter:     updatedTargetMU.CurrentHp,
		DidKoTarget:       didKO,
	})
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("create match turn: %w", err)
	}

	//11. If match completed, update match to complete and commit transaction
	if matchCompleted {
		// finalize match
		_, err = qtx.CompleteMatch(ctx, store.CompleteMatchParams{
			ID: match.ID,
			WinnerPlayerID: sql.NullInt64{
				Int64: winnerID,
				Valid: true,
			},
		})
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("complete match: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit tx: %w", err)
		}
		return nil
	}

	//11. If match not completed
	//--Increment match.CurrentTurnNumber
	currentTurn := match.CurrentTurnNumber
	nextTurn := currentTurn + 1

	//--Decide next actor based on round logic
	var opponentID int64
	if actingPlayerID == match.Player1ID {
		opponentID = match.Player2ID
	} else {
		opponentID = match.Player1ID
	}
	var nextActor int64

	if currentTurn%2 == 1 {
		// first action in this round -> second action goes to opponent
		nextActor = opponentID
	} else {
		// second action in round -> new round, recompute who goes first

		// Recompute speeds from *current* active units (they may have auto-switched)
		// 1) find both sides again (you already have 'sides')
		var p1Side, p2Side *store.MatchSide
		for i := range sides {
			if sides[i].PlayerID == match.Player1ID {
				p1Side = &sides[i]
			} else if sides[i].PlayerID == match.Player2ID {
				p2Side = &sides[i]
			}
		}

		// 2) get active match units for each side
		p1Active, err := qtx.GetActiveMatchUnitForSide(ctx, p1Side.ID)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("get p1 active unit: %w", err)
		}
		p2Active, err := qtx.GetActiveMatchUnitForSide(ctx, p2Side.ID)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("get p2 active unit: %w", err)
		}

		// 3) load units to get BaseSpeed
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

		// 4) compare speeds
		if p1Unit.BaseSpeed > p2Unit.BaseSpeed {
			nextActor = match.Player1ID
		} else if p2Unit.BaseSpeed > p1Unit.BaseSpeed {
			nextActor = match.Player2ID
		} else {
			// tie-breaker random
			if rand.Intn(2) == 0 {
				nextActor = match.Player1ID
			} else {
				nextActor = match.Player2ID
			}
		}
	}

	//12. Update matches with new current_turn_number and current_actor_player_id
	_, err = qtx.UpdateMatchTurnAndActor(ctx, store.UpdateMatchTurnAndActorParams{
		ID:                match.ID,
		CurrentTurnNumber: nextTurn,
		CurrentActorPlayerID: sql.NullInt64{
			Int64: nextActor,
			Valid: true,
		},
	})
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("update match turn/actor: %w", err)
	}
	//13. Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
