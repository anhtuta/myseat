package httpapi

import "net/http"

type healthResponse struct {
	Status string `json:"status"`
}

// health reports process liveness. Postgres and Redis checks are added with those clients.
func health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}
