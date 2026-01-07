package devutil

import (
	"context"
	"database/sql"
)

func ResetDB(ctx context.Context, db *sql.DB) error {
	// Order from most dependent to least dependent
	stmts := []string{
		"TRUNCATE match_units RESTART IDENTITY CASCADE",
		"TRUNCATE match_sides RESTART IDENTITY CASCADE",
		"TRUNCATE match_turns RESTART IDENTITY CASCADE",
		"TRUNCATE matches RESTART IDENTITY CASCADE",
		"TRUNCATE players RESTART IDENTITY CASCADE",
		"TRUNCATE squad_units RESTART IDENTITY CASCADE",
		"TRUNCATE squads RESTART IDENTITY CASCADE",
	}

	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}
