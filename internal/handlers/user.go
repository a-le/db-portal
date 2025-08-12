package handlers

import (
	"db-portal/internal/contextkeys"
	"db-portal/internal/internaldb"
	"db-portal/internal/response"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type userResp = response.Response[[]internaldb.User]

func (s *Services) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	currentUsername := contextkeys.UsernameFromContext(r.Context())
	resp := userResp{}
	var err error

	resp.Data, err = s.Store.GetAllUsers(currentUsername)
	if err != nil {
		resp.Error = err.Error()
		response.WriteJSON(w, http.StatusInternalServerError, &resp)
		return
	}

	response.WriteJSON(w, http.StatusOK, &resp)
}

func (s *Services) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	currentUsername := contextkeys.UsernameFromContext(r.Context())
	resp := response.BasicResponse{}

	// Parse user data from JSON body
	var req struct {
		Username string `json:"username"`
		IsAdmin  string `json:"isadmin"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error = "invalid json. " + err.Error()
		response.WriteJSON(w, http.StatusBadRequest, &resp)
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	isAdminInt, err := strconv.Atoi(req.IsAdmin)
	if err != nil {
		resp.Error = "isadmin must be an integer"
		response.WriteJSON(w, http.StatusBadRequest, &resp)
		return
	}

	err = s.Store.CreateUser(currentUsername, req.Username, isAdminInt, req.Password)
	if err != nil {
		resp.Error = "Cannot add user. " + err.Error()
		response.WriteJSON(w, http.StatusInternalServerError, &resp)
		return
	}

	response.WriteJSON(w, http.StatusOK, &resp)
}
