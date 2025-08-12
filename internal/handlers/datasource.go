package handlers

import (
	"db-portal/internal/contextkeys"
	"db-portal/internal/internaldb"
	"db-portal/internal/response"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type userDataSourceResp = response.Response[[]internaldb.DataSource]

func (s *Services) HandleListUserDataSources(w http.ResponseWriter, r *http.Request) {

	currentUsername := contextkeys.UsernameFromContext(r.Context())
	username := chi.URLParam(r, "username")

	resp := userDataSourceResp{}
	var err error
	resp.Data, err = s.Store.GetAllUserDataSources(currentUsername, username)
	if err != nil {
		resp.Error = err.Error()
		response.WriteJSON(w, http.StatusInternalServerError, &resp)
		return
	}

	response.WriteJSON(w, http.StatusOK, &resp)
}

func (s *Services) HandleListUserAvailableDataSources(w http.ResponseWriter, r *http.Request) {

	currentUsername := contextkeys.UsernameFromContext(r.Context())
	username := chi.URLParam(r, "username")

	resp := userDataSourceResp{}
	var err error
	resp.Data, err = s.Store.GetAllUserAvailableDataSources(currentUsername, username)
	if err != nil {
		resp.Error = err.Error()
		response.WriteJSON(w, http.StatusInternalServerError, &resp)
		return
	}

	response.WriteJSON(w, http.StatusOK, &resp)
}

func (s *Services) HandleCreateDataSource(w http.ResponseWriter, r *http.Request) {
	currentUsername := contextkeys.UsernameFromContext(r.Context())

	resp := userDataSourceResp{}

	// get name, vendor, location from JSON
	var req struct {
		Name     string `json:"name"`
		Vendor   string `json:"vendor"`
		Location string `json:"location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error = "invalid json. " + err.Error()
		response.WriteJSON(w, http.StatusBadRequest, &resp)
		return
	}
	req.Name = strings.TrimSpace(req.Name)

	err := s.Store.CreateDataSource(currentUsername, req.Name, req.Vendor, req.Location)
	if err != nil {
		resp.Error = "cannot create new data source. " + err.Error()
		response.WriteJSON(w, http.StatusInternalServerError, &resp)
		return
	}

	response.WriteJSON(w, http.StatusOK, &resp)
}

func (s *Services) HandleCreateUserDataSource(w http.ResponseWriter, r *http.Request) {
	currentUsername := contextkeys.UsernameFromContext(r.Context())

	resp := userDataSourceResp{}

	username := chi.URLParam(r, "username")
	dsName := chi.URLParam(r, "dsName")

	err := s.Store.CreateUserDataSource(currentUsername, username, dsName)
	if err != nil {
		resp.Error = "cannot add data source to user. " + err.Error()
		response.WriteJSON(w, http.StatusInternalServerError, &resp)
		return
	}

	response.WriteJSON(w, http.StatusOK, &resp)
}

func (s *Services) HandleDeleteUserDataSource(w http.ResponseWriter, r *http.Request) {
	currentUsername := contextkeys.UsernameFromContext(r.Context())

	resp := userDataSourceResp{}

	username := chi.URLParam(r, "username")
	dsName := chi.URLParam(r, "dsName")

	err := s.Store.DeleteUserDataSource(currentUsername, username, dsName)
	if err != nil {
		resp.Error = "cannot delete user data source." + err.Error()
		response.WriteJSON(w, http.StatusInternalServerError, &resp)
		return
	}

	response.WriteJSON(w, http.StatusOK, &resp)
}

func (s *Services) HandleDataSourceTest(w http.ResponseWriter, r *http.Request) {
	currentUsername := contextkeys.UsernameFromContext(r.Context())

	resp := userDataSourceResp{}

	ok, _ := s.Store.CheckIsAdmin(currentUsername)
	if !ok {
		resp.Error = "you are not allowed to test data sources. "
		response.WriteJSON(w, http.StatusBadRequest, &resp)
		return
	}

	// get params from JSON
	var req struct {
		Vendor   string `json:"vendor"`
		Location string `json:"location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error = "invalid json. " + err.Error()
		response.WriteJSON(w, http.StatusBadRequest, &resp)
		return
	}
	req.Vendor = strings.TrimSpace(req.Vendor)

	ok, err := s.Store.TestDataSource(req.Vendor, req.Location)
	if !ok {
		// set resp.Error, http.StatusOK is fine here
		if err != nil {
			resp.Error = err.Error()
		}
	}

	response.WriteJSON(w, http.StatusOK, &resp)
}
