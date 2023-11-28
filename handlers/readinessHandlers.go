package handlers

import "net/http"

func ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	RespondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func ErrorHandler(w http.ResponseWriter, r *http.Request) {
	RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
}
