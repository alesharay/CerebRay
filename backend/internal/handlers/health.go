package handlers

import "net/http"

// HandleHealth returns 200 OK for liveness probes.
func HandleHealth(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
