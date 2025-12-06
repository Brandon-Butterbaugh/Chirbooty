package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Brandon-Butterbaugh/Chirbooty.git/internal/auth"
	"github.com/Brandon-Butterbaugh/Chirbooty.git/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Password     string    `json:"-"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) newUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create hash", err)
		return
	}

	user, err := cfg.database.CreateUser(r.Context(),
		database.CreateUserParams{
			Email:          params.Email,
			HashedPassword: hash},
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	respBody := User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed.Bool,
	}

	respondWithJSON(w, http.StatusCreated, respBody)
}

func (cfg *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {

	// Get access token
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authorization header failed", err)
		return
	}

	// Validate user with token
	id, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "JWT validation failed", err)
		return
	}

	// Get user with id
	user, err := cfg.database.GetUserFromID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get user", err)
		return
	}

	// Read request body
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	// Check for password update, hash new password, update hash to user
	if params.Password != "" {
		hash, err := auth.HashPassword(params.Password)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't create hash", err)
			return
		}
		user, err = cfg.database.UpdateUserPassword(r.Context(),
			database.UpdateUserPasswordParams{
				ID:             user.ID,
				HashedPassword: hash,
			},
		)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't update user password", err)
			return
		}
	}

	// Check for email update, update email to user
	if params.Email != "" {
		user, err = cfg.database.UpdateUserEmail(r.Context(),
			database.UpdateUserEmailParams{
				ID:    user.ID,
				Email: params.Email,
			},
		)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't update user email", err)
			return
		}
	}
	// Find refresh token
	refresh, err := cfg.database.GetRefreshTokenFromUser(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't find refresh token from user", err)
		return
	}

	respBody := User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refresh.Token,
		IsChirpyRed:  user.IsChirpyRed.Bool,
	}

	respondWithJSON(w, http.StatusOK, respBody)
}

func (cfg *apiConfig) login(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.database.GetUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find user", err)
		return
	}

	authorization, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error checking password", err)
		return
	}

	if !authorization {
		respondWithError(w, http.StatusUnauthorized, "incorrect password", err)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error making JWT", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error making refresh token", err)
		return
	}
	refresh, err := cfg.database.CreateRefreshToken(
		r.Context(),
		database.CreateRefreshTokenParams{
			Token:     refreshToken,
			UserID:    user.ID,
			ExpiresAt: time.Now().AddDate(0, 0, 60),
		},
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating refresh token in database", err)
		return
	}

	respBody := User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refresh.Token,
		IsChirpyRed:  user.IsChirpyRed.Bool,
	}

	respondWithJSON(w, http.StatusOK, respBody)
}

func (cfg *apiConfig) refresh(w http.ResponseWriter, r *http.Request) {
	type refreshResponse struct {
		Token string `json:"token"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authorization header failed", err)
		return
	}

	user, err := cfg.database.GetUserFromRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Refresh token is bad/expired/revoked", err)
		return
	}

	newToken, err := auth.MakeJWT(user.ID, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error making JWT", err)
		return
	}

	respBody := refreshResponse{
		Token: newToken,
	}

	respondWithJSON(w, http.StatusOK, respBody)
}

func (cfg *apiConfig) revoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authorization header failed", err)
		return
	}

	err = cfg.database.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error revoking refresh token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) upgrade(w http.ResponseWriter, r *http.Request) {

	// Define data and parameters
	type data struct {
		UserID string `json:"user_id"`
	}
	type parameters struct {
		Event string `json:"event"`
		Data  data   `json:"data"`
	}

	// Unmarshal
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	// Check event
	if params.Event != "user.upgraded" {
		respondWithError(w, http.StatusNoContent, "Wrong event", err)
		return
	}

	// Check APIkey
	apiKey, err := auth.GetAPIKey(r.Header)
	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Wrong APIkey", err)
		return
	}

	// Get user
	id, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Invalid uuid", err)
		return
	}
	user, err := cfg.database.GetUserFromID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find user", err)
		return
	}

	// Update user to red
	_, err = cfg.database.UpdateUserRed(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't upgrade to red", err)
		return
	}

	// Return status code
	w.WriteHeader(http.StatusNoContent)
}
