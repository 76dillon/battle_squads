# battle_squads

#Access postgres
psql postgres

#Access Database
\c DATABASE_NAME

#Database Migrations
goose postgres "postgres://mattdillon@localhost:5432/battle_squads?sslmode=disable" up/down

#Generate SQLC 
sqlc generate

