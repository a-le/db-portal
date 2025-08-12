package handlers

import (
	"db-portal/internal/contextkeys"
	"db-portal/internal/dbutil"
	"db-portal/internal/response"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Services) HandleUserDataSourceTest(w http.ResponseWriter, r *http.Request) {

	currentUsername := contextkeys.UsernameFromContext(r.Context())
	dsName := chi.URLParam(r, "dsName")
	resp := response.BasicResponse{}

	// get ds info from internal DB
	ds, err := s.Store.RequireUserDataSource(currentUsername, currentUsername, dsName)
	if err != nil {
		resp.Error = err.Error()
		response.WriteJSON(w, http.StatusInternalServerError, &resp)
		return
	}

	// try to get conn from DB server
	conn, err := dbutil.GetConn(ds.Vendor, ds.Location, false)
	if err != nil {
		resp.Error = err.Error()
		response.WriteJSON(w, http.StatusInternalServerError, &resp)
		return
	}
	conn.Close()

	// send response
	response.WriteJSON(w, http.StatusOK, &resp)
}
