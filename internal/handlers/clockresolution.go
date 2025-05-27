package handlers

import (
	"db-portal/internal/response"
	"db-portal/internal/timer"
	"net/http"
)

func (s *Services) ClockResolutionHandler(w http.ResponseWriter, r *http.Request) {
	// Cache the value after first computation
	if s.clockResolution == 0 {
		s.clockResolution = timer.EstimateMinClockResolution(10000)
	}

	response.SendJSON(&response.Data{
		Data: s.clockResolution,
	}, w)
}
