package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Chinchzilla/chirpy/internal/auth"
	"github.com/google/uuid"
)

const (
	UserUpgradedEvent string = "user.upgraded"
)

type WebhookRequest struct {
	Event string `json:"event"`
	Data  struct {
		UserID uuid.UUID `json:"user_id"`
	} `json:"data"`
}

func (cfg *apiConfig) handlerEvent(w http.ResponseWriter, r *http.Request) {
	receivedToken, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No token provided in the header", err)
		return
	}

	if cfg.polkaKey != receivedToken {
		respondWithError(w, http.StatusUnauthorized, "Provided token is invalid", err)
		return
	}

	req := WebhookRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode the json body", err)
		return
	}

	switch req.Event {
	case UserUpgradedEvent:
		if err := cfg.dbQueries.UpdateUserToChirpyRed(r.Context(), req.Data.UserID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respondWithError(w, http.StatusNotFound, "Couldn't find the User by ID", err)
				return
			}
			respondWithError(w, http.StatusInternalServerError, "Unknown error when trying to upgrade user to chirpy red", err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	default:
		w.WriteHeader(http.StatusNoContent)
		return
	}
}
