package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Brandon-Butterbaugh/Chirbooty.git/internal/auth"
	"github.com/Brandon-Butterbaugh/Chirbooty.git/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	var respBody []Chirp
	authorID := r.URL.Query().Get("author_id")
	sortType := r.URL.Query().Get("sort")
	var chirps []database.Chirp
	var err error

	if authorID != "" {
		id, err := uuid.Parse(authorID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "Invalid uuid", err)
			return
		}
		chirps, err = cfg.database.GetAuthorChirps(r.Context(), id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps", err)
			return
		}
	} else {
		chirps, err = cfg.database.GetChirps(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps", err)
			return
		}
	}

	if sortType == "desc" {
		sort.Slice(chirps,
			func(i, j int) bool { return chirps[i].CreatedAt.After(chirps[j].CreatedAt) },
		)
	}

	for _, chirp := range chirps {
		respBody = append(respBody, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		},
		)
	}

	respondWithJSON(w, http.StatusOK, respBody)
}

func (cfg *apiConfig) getChirp(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Invalid uuid", err)
		return
	}

	chirp, err := cfg.database.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp", err)
		return
	}

	respBody := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	respondWithJSON(w, http.StatusOK, respBody)
}

func (cfg *apiConfig) newChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authorization header failed", err)
		return
	}
	id, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "JWT validation failed", err)
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	cleanBody := cleanProfanity(params.Body)

	chirp, err := cfg.database.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanBody,
		UserID: id,
	},
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
		return
	}

	respBody := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	respondWithJSON(w, http.StatusCreated, respBody)
}

func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {
	// Authenticate user
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authorization header failed", err)
		return
	}
	id, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "JWT validation failed", err)
		return
	}

	// Get the chirp
	chirpIDString := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Invalid uuid", err)
		return
	}

	chirp, err := cfg.database.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp", err)
		return
	}

	// Check user is chirp author
	if chirp.UserID != id {
		respondWithError(w, http.StatusForbidden, "Not the chirp author", err)
		return
	}

	// Delete the chirp
	err = cfg.database.DeleteChirp(r.Context(), chirp.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete chirp", err)
		return
	}

	// Return status code
	w.WriteHeader(http.StatusNoContent)
}

func cleanProfanity(str string) string {
	temp := str
	temp = strings.ToLower(temp)
	split := strings.Split(temp, " ")
	strSplit := strings.Split(str, " ")

	for i, word := range split {
		if word == "kerfuffle" || word == "sharbert" || word == "fornax" {
			strSplit[i] = "****"
		}
	}
	return strings.Join(strSplit, " ")
}
