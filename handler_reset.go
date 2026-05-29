package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if err := cfg.dbQeries.WipeUsers(r.Context()); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't wipe the users table", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Wiped the users database")
}
