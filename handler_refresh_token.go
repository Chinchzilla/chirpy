package main

import (
	"net/http"
	"time"

	"github.com/Chinchzilla/chirpy/internal/auth"
)

type AccessToken struct {
	Token string `json:"token"`
}

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't find the token in the header", err)
		return
	}

	lookup, err := cfg.dbQueries.GetUserFromRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't fetch user by refresh token or the token is expired", err)
		return
	}

	if lookup.RevokedAt.Valid || time.Now().After(lookup.ExpiresAt) {
		respondWithError(w, http.StatusUnauthorized, "Refresh token for the user is expired", nil)
		return
	}

	accessToken, err := auth.MakeJWT(lookup.UserID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate access token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, &AccessToken{Token: accessToken})
}

func (cfg *apiConfig) handlerRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get the token from the header", err)
		return
	}

	err = cfg.dbQueries.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error while trying to revoke the refresh token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
