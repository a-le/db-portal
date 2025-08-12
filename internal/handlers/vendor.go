package handlers

import (
	"db-portal/internal/internaldb"
	"db-portal/internal/response"
	"net/http"
)

type vendorResponse = response.Response[[]internaldb.Vendor]

func (s *Services) HandleListVendors(w http.ResponseWriter, r *http.Request) {

	resp := vendorResponse{}
	var err error
	resp.Data, err = s.Store.GetAllVendors()
	if err != nil {
		resp.Error = err.Error()
		response.WriteJSON(w, http.StatusInternalServerError, &resp)
		return
	}

	response.WriteJSON(w, http.StatusOK, &resp)
}
