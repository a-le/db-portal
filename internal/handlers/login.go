package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"db-portal/internal/response"
	"db-portal/internal/security"

	"github.com/golang-jwt/jwt/v5"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

func (s *Services) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	var resp loginResponse

	// Parse JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error = "Invalid request."
		response.WriteJSON(w, http.StatusBadRequest, &resp)
		return
	}

	// Check credentials
	ok, user, err := s.Store.CheckUserCredentials(req.Username, req.Password)
	if err != nil {
		resp.Error = "Internal error."
		response.WriteJSON(w, http.StatusInternalServerError, &resp)
		return
	}
	if !ok {
		resp.Error = "Invalid username or password."
		response.WriteJSON(w, http.StatusUnauthorized, &resp)
		return
	}

	// Generate JWT token
	claims := jwt.MapClaims{
		"sub":     req.Username,
		"name":    user.Name,
		"isadmin": user.IsAdmin,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(security.JWTSecretKey)
	if err != nil {
		resp.Error = "Failed to generate token."
		response.WriteJSON(w, http.StatusInternalServerError, &resp)
		return
	}

	resp.Token = tokenString
	response.WriteJSON(w, http.StatusOK, &resp)
}
