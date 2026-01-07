package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/76dillon/battle_squads/internal/game"
	"github.com/76dillon/battle_squads/internal/store"
)

type Server struct {
	mux *http.ServeMux
	q   *store.Queries
	svc *game.Service
}

type postTurnRequest struct {
	MoveID int64 `json:"move_id"`
}

func NewServer(q *store.Queries, svc *game.Service) *Server {
	s := &Server{
		mux: http.NewServeMux(),
		q:   q,
		svc: svc,
	}

	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/login", s.handleLogin)
	s.mux.HandleFunc("/units", s.handleListUnits)
	s.mux.HandleFunc("/me/squads", s.handleSquads)
	s.mux.HandleFunc("/matches/", s.handleMatch)
	s.mux.HandleFunc("/me/matches", s.handleListMyMatches)
	s.mux.HandleFunc("/matches", s.handleCreateMatch) // for POST /matches
	// we'll add /matches/{id} soon
}

func (s *Server) handleMatch(w http.ResponseWriter, r *http.Request) {
	// URL is something like:
	//   /matches/1
	//   /matches/1/turns

	// Strip the prefix
	path := strings.TrimPrefix(r.URL.Path, "/matches/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	// POST /matches/{id}/turns
	if r.Method == http.MethodPost && strings.HasSuffix(path, "/turns") {
		s.handlePostTurn(w, r)
		return
	}

	// GET /matches/{id}
	if r.Method == http.MethodGet && !strings.Contains(path, "/") {
		s.handleGetMatch(w, r)
		return
	}

	http.NotFound(w, r)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Simple CORS for local dev
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8000")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Player-ID")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleGetMatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// URL: /matches/{id}
	// strip prefix and parse id
	idStr := strings.TrimPrefix(r.URL.Path, "/matches/")
	if idStr == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	matchID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	match, err := s.q.GetMatchByID(ctx, matchID)
	if err != nil {
		// ideally check for sql.ErrNoRows
		http.Error(w, "match not found", http.StatusNotFound)
		return
	}

	sides, err := s.q.GetMatchSidesByMatchID(ctx, matchID)
	if err != nil {
		http.Error(w, "error loading sides", http.StatusInternalServerError)
		return
	}

	unitsBySide := make(map[int64][]store.MatchUnit, len(sides))
	for _, side := range sides {
		units, err := s.q.GetMatchUnitsBySideID(ctx, side.ID)
		if err != nil {
			http.Error(w, "error loading units", http.StatusInternalServerError)
			return
		}
		unitsBySide[side.ID] = units
	}

	resp := s.buildMatchResponse(match, sides, unitsBySide)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "encode error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handlePostTurn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// parse match ID from URL: /matches/{id}/turns
	path := strings.TrimPrefix(r.URL.Path, "/matches/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 || parts[1] != "turns" {
		http.Error(w, "bad path", http.StatusBadRequest)
		return
	}
	matchID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "invalid match id", http.StatusBadRequest)
		return
	}

	// TEMP: acting player from header
	playerIDStr := r.Header.Get("X-Player-ID")
	if playerIDStr == "" {
		http.Error(w, "missing X-Player-ID header", http.StatusBadRequest)
		return
	}
	actingPlayerID, err := strconv.ParseInt(playerIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid X-Player-ID", http.StatusBadRequest)
		return
	}

	// decode JSON body
	var req postTurnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.MoveID == 0 {
		http.Error(w, "move_id is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// apply turn
	if err := s.svc.ApplyTurn(ctx, matchID, actingPlayerID, req.MoveID); err != nil {
		switch e := err.(type) {
		case game.ErrWrongTurn:
			http.Error(w, e.Error(), http.StatusConflict) // 409
			return
		case game.ErrMatchNotInProgress:
			http.Error(w, e.Error(), http.StatusBadRequest)
			return
		case game.ErrIllegalMove:
			http.Error(w, e.Error(), http.StatusBadRequest)
			return
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	// load updated match view (reuse logic from handleGetMatch)
	match, err := s.q.GetMatchByID(ctx, matchID)
	if err != nil {
		http.Error(w, "match not found after turn", http.StatusInternalServerError)
		return
	}

	sides, err := s.q.GetMatchSidesByMatchID(ctx, matchID)
	if err != nil {
		http.Error(w, "error loading sides", http.StatusInternalServerError)
		return
	}

	unitsBySide := make(map[int64][]store.MatchUnit, len(sides))
	for _, side := range sides {
		units, err := s.q.GetMatchUnitsBySideID(ctx, side.ID)
		if err != nil {
			http.Error(w, "error loading units", http.StatusInternalServerError)
			return
		}
		unitsBySide[side.ID] = units
	}

	resp := s.buildMatchResponse(match, sides, unitsBySide)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "encode error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleListMyMatches(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	playerIDStr := r.Header.Get("X-Player-ID")
	if playerIDStr == "" {
		http.Error(w, "missing X-Player-ID", http.StatusBadRequest)
		return
	}

	playerID, err := strconv.ParseInt(playerIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid X-Player-ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	matches, err := s.q.ListMatchesForPlayer(ctx, playerID)
	if err != nil {
		http.Error(w, "could not list matches", http.StatusInternalServerError)
		return
	}

	// You can reuse MatchView here
	out := make([]MatchView, 0, len(matches))
	for _, m := range matches {
		// convert to MatchView like in buildMatchResponse
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
		var actor *int64
		if m.CurrentActorPlayerID.Valid {
			actor = &m.CurrentActorPlayerID.Int64
		}

		out = append(out, MatchView{
			ID:                   m.ID,
			State:                MatchState(m.State),
			CreatedAt:            m.CreatedAt,
			StartedAt:            startedAt,
			CompletedAt:          completedAt,
			Player1ID:            m.Player1ID,
			Player2ID:            m.Player2ID,
			WinnerPlayerID:       winnerID,
			CurrentTurnNumber:    int(m.CurrentTurnNumber),
			CurrentActorPlayerID: actor,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

type createMatchRequest struct {
	OpponentPlayerID int64 `json:"opponent_player_id"`
	Player1SquadID   int64 `json:"player1_squad_id"`
	Player2SquadID   int64 `json:"player2_squad_id"`
}

func (s *Server) handleCreateMatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	playerIDStr := r.Header.Get("X-Player-ID")
	if playerIDStr == "" {
		http.Error(w, "missing X-Player-ID", http.StatusBadRequest)
		return
	}
	playerID, err := strconv.ParseInt(playerIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid X-Player-ID", http.StatusBadRequest)
		return
	}

	var req createMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Create match with sqlc
	m, err := s.q.CreateMatch(ctx, store.CreateMatchParams{
		Player1ID: playerID,
		Player2ID: req.OpponentPlayerID,
	})
	if err != nil {
		http.Error(w, "could not create match", http.StatusInternalServerError)
		return
	}

	// Start match using chosen squads
	if err := s.svc.StartMatch(ctx, m.ID, req.Player1SquadID, req.Player2SquadID); err != nil {
		http.Error(w, "could not start match", http.StatusInternalServerError)
		return
	}

	// Load full view and return
	match, err := s.q.GetMatchByID(ctx, m.ID)
	if err != nil {
		http.Error(w, "match not found after create", http.StatusInternalServerError)
		return
	}

	sides, err := s.q.GetMatchSidesByMatchID(ctx, m.ID)
	if err != nil {
		http.Error(w, "error loading sides", http.StatusInternalServerError)
		return
	}

	unitsBySide := make(map[int64][]store.MatchUnit, len(sides))
	for _, side := range sides {
		units, err := s.q.GetMatchUnitsBySideID(ctx, side.ID)
		if err != nil {
			http.Error(w, "error loading units", http.StatusInternalServerError)
			return
		}
		unitsBySide[side.ID] = units
	}

	resp := s.buildMatchResponse(match, sides, unitsBySide) // if buildMatchResponse is a method

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// GET /units
func (s *Server) handleListUnits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	units, err := s.q.ListUnits(ctx)
	if err != nil {
		http.Error(w, "could not list units", http.StatusInternalServerError)
		return
	}

	// Simple DTO
	type UnitDTO struct {
		ID         int64  `json:"id"`
		Name       string `json:"name"`
		TypeID     int64  `json:"type_id"`
		BaseHP     int32  `json:"base_hp"`
		BaseAttack int32  `json:"base_attack"`
		BaseSpeed  int32  `json:"base_speed"`
	}

	out := make([]UnitDTO, 0, len(units))
	for _, u := range units {
		out = append(out, UnitDTO{
			ID:         u.ID,
			Name:       u.Name,
			TypeID:     u.TypeID,
			BaseHP:     u.BaseHp,
			BaseAttack: u.BaseAttack,
			BaseSpeed:  u.BaseSpeed,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (s *Server) handleSquads(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListMySquads(w, r)
	case http.MethodPost:
		s.handleCreateSquad(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleListMySquads(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Header.Get("X-Player-ID")
	if playerIDStr == "" {
		http.Error(w, "missing X-Player-ID", http.StatusBadRequest)
		return
	}
	playerID, err := strconv.ParseInt(playerIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid X-Player-ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	squads, err := s.q.GetSquadsForPlayer(ctx, playerID)
	if err != nil {
		http.Error(w, "could not list squads", http.StatusInternalServerError)
		return
	}

	type SquadDTO struct {
		ID    int64   `json:"id"`
		Name  string  `json:"name"`
		Units []int64 `json:"units"` // unit IDs in order
	}

	out := make([]SquadDTO, 0, len(squads))
	for _, sq := range squads {
		sus, err := s.q.GetSquadUnits(ctx, sq.ID)
		if err != nil {
			http.Error(w, "could not load squad units", http.StatusInternalServerError)
			return
		}
		units := make([]int64, len(sus))
		for i, su := range sus {
			units[i] = su.UnitID
		}
		out = append(out, SquadDTO{
			ID:    sq.ID,
			Name:  sq.Name,
			Units: units,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

type createSquadRequest struct {
	Name    string  `json:"name"`
	UnitIDs []int64 `json:"unit_ids"`
}

func (s *Server) handleCreateSquad(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.Header.Get("X-Player-ID")
	if playerIDStr == "" {
		http.Error(w, "missing X-Player-ID", http.StatusBadRequest)
		return
	}
	playerID, err := strconv.ParseInt(playerIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid X-Player-ID", http.StatusBadRequest)
		return
	}

	var req createSquadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Name == "" || len(req.UnitIDs) == 0 {
		http.Error(w, "name and unit_ids are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Create squad
	sq, err := s.q.CreateSquad(ctx, store.CreateSquadParams{
		PlayerID: playerID,
		Name:     req.Name,
	})
	if err != nil {
		http.Error(w, "could not create squad", http.StatusInternalServerError)
		return
	}

	// Create squad_units in order
	for pos, unitID := range req.UnitIDs {
		_, err := s.q.CreateSquadUnit(ctx, store.CreateSquadUnitParams{
			SquadID:  sq.ID,
			UnitID:   unitID,
			Position: int32(pos),
		})
		if err != nil {
			http.Error(w, "could not create squad units", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}
