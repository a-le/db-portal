package handlers

import (
	"db-portal/internal/response"
	"db-portal/internal/timer"
	"net/http"
	"time"
)

type clockResolutionResp = response.Response[time.Duration]

func (s *Services) HandleClockResolution(w http.ResponseWriter, r *http.Request) {
	resp := clockResolutionResp{}
	// Cache the value after first computation
	if s.clockResolution == 0 {
		s.clockResolution = timer.EstimateMinClockResolution(10000)
	}
	resp.Data = s.clockResolution

	response.WriteJSON(w, http.StatusOK, &resp)
}
