package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/76dillon/battle_squads/internal/config"
	"github.com/76dillon/battle_squads/internal/game"
	httpapi "github.com/76dillon/battle_squads/internal/http"
	"github.com/76dillon/battle_squads/internal/store"
	_ "github.com/lib/pq" // register postgres driver
)

func main() {

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
	//ctx := context.Background()
	//if err := devutil.ResetDB(ctx, db); err != nil {
	//	fmt.Fprintf(os.Stderr, "error resetting db: %v\n", err)
	//	os.Exit(1)
	//}

	// 7. Start HTTP server on cfg.HTTPPort
	api := httpapi.NewServer(q, svc)

	addr := ":" + cfg.HTTPPort
	fmt.Println("listening on", addr)
	if err := http.ListenAndServe(addr, api); err != nil {
		fmt.Fprintf(os.Stderr, "server error %v\n", err)
		os.Exit(1)
	}
}
