package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/76dillon/battle_squads/internal/store"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	PlayerID int64  `json:"player_id"`
	Username string `json:"username"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	p, err := s.q.GetPlayerByUsername(ctx, req.Username)
	if err != nil {
		// you might check for sql.ErrNoRows
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// For now, assume password_hash stores plain text (dev only!)
	if req.Password != p.PasswordHash {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	resp := loginResponse{
		PlayerID: p.ID,
		Username: p.Username,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

type signupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type signupResponse struct {
	PlayerID int64  `json:"player_id"`
	Username string `json:"username"`
}

func (s *Server) handleSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	// DEV ONLY: store plain password, or hash with bcrypt if you like
	ctx := r.Context()
	p, err := s.q.CreatePlayer(ctx, store.CreatePlayerParams{
		Username:     req.Username,
		PasswordHash: req.Password, // or bcrypt hash
	})
	if err != nil {
		// you can check for unique violation here if needed
		http.Error(w, "could not create user", http.StatusBadRequest)
		return
	}

	resp := signupResponse{
		PlayerID: p.ID,
		Username: p.Username,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) requireAdmin(ctx context.Context, playerID int64) (*store.Player, error) {
	p, err := s.q.GetPlayerByID(ctx, playerID)
	if err != nil {
		return nil, err
	}
	if !p.IsAdmin {
		return nil, fmt.Errorf("not admin")
	}
	return &p, nil
}
