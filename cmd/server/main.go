package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/76dillon/battle_squads/internal/config"
	"github.com/76dillon/battle_squads/internal/devutil"
	"github.com/76dillon/battle_squads/internal/game"
	httpapi "github.com/76dillon/battle_squads/internal/http"
	"github.com/76dillon/battle_squads/internal/store"
	_ "github.com/lib/pq" // register postgres driver
)

func main() {
	ctx := context.Background()
	// 1. Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading config: %v\n", err)
		os.Exit(1)
	}
	// 2. Open DB: sql.Open("postgres", cfg.DatabaseURL)
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error connecting to db: %v", err)
		os.Exit(1)
	}

	// 3. Ping DB to verify connection
	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "error pinging db: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("successfully connected to database")

	// 4. Defer db.Close()
	defer db.Close()

	// 5. Create store := store.New(db)
	q := store.New(db) // or whatever constructor sqlc generated

	// 6. Create Game Service
	svc := game.NewService(db)

	// For Dev purposes only - delete when done!!!
	if err := devutil.ResetDB(ctx, db); err != nil {
		fmt.Fprintf(os.Stderr, "error resetting db: %v\n", err)
		os.Exit(1)
	}

	//Test StartMatch - Delete when done
	// 1) Create two players
	p1, _ := q.CreatePlayer(ctx, store.CreatePlayerParams{
		Username:     "alice",
		PasswordHash: "hash1",
	})
	p2, _ := q.CreatePlayer(ctx, store.CreatePlayerParams{
		Username:     "bob",
		PasswordHash: "hash2",
	})

	// 2) Create squads for each player
	s1, _ := q.CreateSquad(ctx, store.CreateSquadParams{
		PlayerID: p1.ID,
		Name:     "Alice Squad",
	})
	s2, _ := q.CreateSquad(ctx, store.CreateSquadParams{
		PlayerID: p2.ID,
		Name:     "Bob Squad",
	})

	// attach some units (use IDs that exist from your seed: 1,2,3, etc.)
	_, _ = q.CreateSquadUnit(ctx, store.CreateSquadUnitParams{
		SquadID:  s1.ID,
		UnitID:   1,
		Position: 0,
	})
	_, _ = q.CreateSquadUnit(ctx, store.CreateSquadUnitParams{
		SquadID:  s1.ID,
		UnitID:   2,
		Position: 1,
	})
	_, _ = q.CreateSquadUnit(ctx, store.CreateSquadUnitParams{
		SquadID:  s1.ID,
		UnitID:   3,
		Position: 2,
	})

	// same for s2
	_, _ = q.CreateSquadUnit(ctx, store.CreateSquadUnitParams{
		SquadID:  s2.ID,
		UnitID:   1,
		Position: 0,
	})

	_, _ = q.CreateSquadUnit(ctx, store.CreateSquadUnitParams{
		SquadID:  s2.ID,
		UnitID:   2,
		Position: 1,
	})
	_, _ = q.CreateSquadUnit(ctx, store.CreateSquadUnitParams{
		SquadID:  s2.ID,
		UnitID:   3,
		Position: 2,
	})

	// 3) Create match
	//m, _ := q.CreateMatch(ctx, store.CreateMatchParams{
	//	Player1ID: p1.ID,
	//	Player2ID: p2.ID,
	//})

	// 4) Call StartMatch
	//if err := svc.StartMatch(ctx, m.ID, s1.ID, s2.ID); err != nil {
	//	fmt.Fprintf(os.Stderr, "error starting match: %v\n", err)
	//	os.Exit(1)
	//}

	// 5) Fetch and print some state
	//m2, _ := q.GetMatchByID(ctx, m.ID)
	//fmt.Printf("match after StartMatch: %+v\n", m2)

	//sides, _ := q.GetMatchSidesByMatchID(ctx, m.ID)
	//fmt.Printf("sides: %+v\n", sides)

	//for _, side := range sides {
	//	units, _ := q.GetMatchUnitsBySideID(ctx, side.ID)
	//	fmt.Printf("side %d units: %+v\n", side.ID, units)
	//}

	//End game service test queries

	// 7. Start HTTP server on cfg.HTTPPort
	api := httpapi.NewServer(q, svc)

	addr := ":" + cfg.HTTPPort
	fmt.Println("listening on", addr)
	if err := http.ListenAndServe(addr, api); err != nil {
		fmt.Fprintf(os.Stderr, "server error %v\n", err)
		os.Exit(1)
	}
}
