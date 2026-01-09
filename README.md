# Battle Squads

## Description

The Battle Squads web app is inspired by Pokemon and enables users to play each other using their squads of units (pokemon). Players can have up to three units in a squad and users' squads will face off in a battle match.

![Battle Squads Web UI](assets/app_pic.png "Battle Squads Web UI")

## Motivation

Growing up, I've enjoyed playing Pokemon and was looking to create an application that enables users to battle each other with Pokemon. I wanted to integrate this into an HTTP API for users to utilize the endpoints to create units, squads, matches, and battle. The open source nature enables users to add custom units and movesets to the database to make it their own.

## Quick Start

1. Clone the repo: ```git clone https://github.com/76dillon/battle_squads.git```
2. Install Go, Postgres, Goose, and SQLC
3. On the Server Side - create the database and use Goose to migrate the database
    1. ```psql postgres```
    2. ```CREATE DATABASE battle_squads``` (Database can be accessed at anytime with \c DB_NAME)
    3. From the root of the battle squads directory: ```cd sql/schema```
    4. Run the up migration: ```goose postgres "postgres://username:password@localhost:5432/battle_squads?sslmode=disable" up``` Should migrate up to v13. 
4. Create an env file in the root of the working directory: ```touch .env```
5. Copy the following lines of code, modifying the username and password of your postgres database: 
```
export DB_URL="postgres://username:password@localhost:5432/battle_squads?sslmode=disable"
export HTTP_PORT="8080"
```
6. Run the server at the root of the working directory: ```go run ./cmd/server```
7. Run the web client in the web directory ```python3 -m http.server 8000```
8. Open a browser and navigate to ```http://localhost:8000```
