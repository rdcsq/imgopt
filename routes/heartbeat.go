package routes

import (
	"encoding/json"
	"net/http"
)

func Heartbeat(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
