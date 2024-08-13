package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Ace0314/Chirpy/internal/auth"
)

type Chirp struct {
	ID       int    `json:"id"`
	AuthorID int    `json:"author_id"`
	Body     string `json:"body"`
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}

	subject, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT")
		return
	}

	userID, err := strconv.Atoi(subject)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse ID")
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	cleaned, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	chirp, err := cfg.DB.CreateChirp(cleaned, userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:       chirp.ID,
		AuthorID: chirp.AuthorID,
		Body:     chirp.Body,
	})
}

func validateChirp(body string) (string, error) {
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}

	cleanedBody := replaceProfane(body)
	return cleanedBody, nil
}

func replaceProfane(msg string) string {
	profane := []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}
	cleaned := strings.Split(msg, " ")
	words := strings.Split(strings.ToLower(msg), " ")
	for i, word := range words {
		for _, worst := range profane {
			if word == worst {
				cleaned[i] = strings.Replace(word, word, "****", 1)
			}
		}
	}
	return strings.Join(cleaned, " ")
}
