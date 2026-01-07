package game

import (
	"database/sql"

	"github.com/76dillon/battle_squads/internal/store"
)

type Service struct {
	db *sql.DB
	q  *store.Queries
}

func NewService(db *sql.DB) *Service {
	return &Service{
		db: db,
		q:  store.New(db), // or whatever sqlc constructor is
	}
}
