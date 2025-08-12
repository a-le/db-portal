package response

import (
	"encoding/json"
	"net/http"
)

type Response[T any] struct {
	Data  T      `json:"data"`
	Error string `json:"error"`
}

type BasicResponse = Response[any]

func WriteJSON(w http.ResponseWriter, status int, resp any) {
	w.Header().Set("Content-Type", "application/json")

	b, err := json.Marshal(resp)
	if err != nil {
		// error calling MarshalJSON
		// set a new resp and status
		status = http.StatusInternalServerError
		resp = BasicResponse{Error: err.Error()}
		b, _ = json.Marshal(resp)
	}

	w.WriteHeader(status)
	w.Write(b)
}
