package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Chinchzilla/chirpy/internal/auth"
	"github.com/Chinchzilla/chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
	AccessToken  *string   `json:"token,omitempty"`
	RefreshToken *string   `json:"refresh_token,omitempty"`
}

func (cfg *apiConfig) handlerAddUser(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	req := request{}
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode the request body", err)
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPassword,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "User creation failed", err)
		return
	}

	jsonCapableUser := User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		IsChirpyRed:  user.IsChirpyRed,
		AccessToken:  nil,
		RefreshToken: nil,
	}

	respondWithJSON(w, http.StatusCreated, &jsonCapableUser)

}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	req := request{}
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode the request body", err)
		return
	}

	existingUser, err := cfg.dbQueries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "User not found", err)
		return
	}

	isPasswordValid, err := auth.CheckPasswordHash(req.Password, existingUser.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't check password hash", err)
		return
	}

	if !isPasswordValid {
		respondWithError(w, http.StatusUnauthorized, "Invalid password", nil)
		return
	}

	accessToken, err := auth.MakeJWT(existingUser.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't make JWT", err)
		return
	}

	refreshToken, err := cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:  auth.MakeRefreshToken(),
		UserID: existingUser.ID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't make refresh token", err)
		return
	}

	jsonCapableUser := User{
		ID:           existingUser.ID,
		CreatedAt:    existingUser.CreatedAt,
		UpdatedAt:    existingUser.UpdatedAt,
		Email:        existingUser.Email,
		IsChirpyRed:  existingUser.IsChirpyRed,
		AccessToken:  &accessToken,
		RefreshToken: &refreshToken.Token,
	}

	respondWithJSON(w, http.StatusOK, &jsonCapableUser)
}

func (cfg *apiConfig) handlerChangePassword(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode request body", err)
		return
	}

	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't extract the token from the header", err)
		return
	}

	userIDByToken, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get user by access token", err)
		return
	}

	passwordHash, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash the password", err)
		return
	}

	updatedUser, err := cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{ID: userIDByToken, Email: params.Email, HashedPassword: passwordHash})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update the user's passowrd", err)
		return
	}

	jsonCapableUser := User{
		ID:           updatedUser.ID,
		CreatedAt:    updatedUser.CreatedAt,
		UpdatedAt:    updatedUser.UpdatedAt,
		Email:        updatedUser.Email,
		IsChirpyRed:  updatedUser.IsChirpyRed,
		AccessToken:  nil,
		RefreshToken: nil,
	}

	respondWithJSON(w, http.StatusOK, &jsonCapableUser)

}
