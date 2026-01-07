package httpapi

import (
	"encoding/json"
	"net/http"
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
