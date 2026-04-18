package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rinjold/go-etl-studio/internal/services"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func Login(svc *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		token, err := svc.Login(r.Context(), req.Email, req.Password)
		if errors.Is(err, services.ErrBadCredentials) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"token": token})
	}
}

func Register(svc *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req registerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		role := req.Role
		if role == "" {
			role = "viewer"
		}

		user, err := svc.Register(r.Context(), req.Email, req.Password, role)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		})
	}
}
