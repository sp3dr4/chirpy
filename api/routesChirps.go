package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"sort"
	"strconv"

	"github.com/sp3dr4/chirpy/internal/entities"
)

func (cfg *apiConfig) handlerListChirps(w http.ResponseWriter, req *http.Request) {
	chirps, err := cfg.db.GetChirps()
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	sort.Slice(chirps, func(i, j int) bool { return chirps[i].Id < chirps[j].Id })
	respondWithJSON(w, 200, chirps)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, req *http.Request) {
	chirpId, err := strconv.Atoi(req.PathValue("chirpId"))
	if err != nil {
		respondWithError(w, 400, "invalid integer for chirp id")
		return
	}
	chirps, err := cfg.db.GetChirps()
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	i := slices.IndexFunc(chirps, func(c entities.Chirp) bool {
		return c.Id == chirpId
	})
	if i == -1 {
		respondWithError(w, 404, "chirp not found")
		return
	}
	respondWithJSON(w, 200, chirps[i])
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, req *http.Request) {
	userId, err := cfg.isAuthenticated(req)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	type request struct {
		Body string `json:"body"`
	}
	chirpReq := request{}
	if err := json.NewDecoder(req.Body).Decode(&chirpReq); err != nil {
		respondWithError(w, 400, "error decoding request body")
		return
	}
	cleaned, err := entities.ValidateChirp(chirpReq.Body)
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}
	chirp, err := cfg.db.CreateChirp(userId, cleaned)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	respondWithJSON(w, 201, chirp)
}
