package response

import (
	"db-portal/internal/db"
	"encoding/json"
	"fmt"
	"net/http"
)

type Data struct {
	Data any `json:"data"`
}

// Define an interface that includes accepted types
type JSONResponse interface {
	db.DResult | Data
}

// SendJSON converts the response to JSON and sends it
func SendJSON[T JSONResponse](response *T, w http.ResponseWriter) {
	// Convert the response to JSON
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("%+v\n", response)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}
