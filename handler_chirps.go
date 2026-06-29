package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/Chinchzilla/chirpy/internal/auth"
	"github.com/Chinchzilla/chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerPostChrip(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode requrest body", err)
		return
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Bearer token missing or invalid", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized request", err)
		return
	}

	chirp, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp", err)
		return
	}

	chirpEntry, err := cfg.dbQueries.AddChirp(r.Context(), database.AddChirpParams{
		Body:   chirp,
		UserID: userID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't add chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirpEntry.ID,
		CreatedAt: chirpEntry.CreatedAt,
		UpdatedAt: chirpEntry.UpdatedAt,
		Body:      chirpEntry.Body,
		UserID:    chirpEntry.UserID,
	})

}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	authorIDStr := r.URL.Query().Get("author_id")

	var chirps []database.Chirp
	var err error

	if authorIDStr == "" {
		chirps, err = cfg.dbQueries.GetAllChirps(r.Context())
	} else {
		var authorID uuid.UUID
		authorID, err = uuid.Parse(authorIDStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author_id", err)
			return
		}
		chirps, err = cfg.dbQueries.GetAllChirpsByUser(r.Context(), authorID)
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps", err)
		return
	}

	sortQuery := r.URL.Query().Get("sort")
	_, isSort := r.URL.Query()["sort"]
	if isSort && sortQuery == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
	} else if isSort {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
		})
	}

	jsonifyChirps := make([]Chirp, len(chirps))
	for i, chirp := range chirps {
		jsonifyChirps[i] = Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
	}

	respondWithJSON(w, http.StatusOK, jsonifyChirps)
}

func (cfg *apiConfig) handlerChripByID(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")

	chirp, err := cfg.dbQueries.GetChirpByID(r.Context(), uuid.MustParse(chirpID))
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})

}

func (cfg *apiConfig) handlerDeleteChrip(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Chirp ID is not and UUID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't extract token from the header", err)
		return
	}

	userIDByToken, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Access denied", err)
		return
	}

	chirp, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get the chirp by provided ID", err)
		return
	}

	if chirp.UserID.String() != userIDByToken.String() {
		respondWithError(w, http.StatusForbidden, "Deleting other user's chirp is forbidden!", nil)
		return
	}

	err = cfg.dbQueries.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete the chirp", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
