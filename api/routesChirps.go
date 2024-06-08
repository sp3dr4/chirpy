package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"strconv"

	"github.com/sp3dr4/chirpy/internal/entities"
)

var errNotFound = fmt.Errorf("chirp not found")

func (cfg *apiConfig) findChirpById(id int) (*entities.Chirp, error) {
	chirps, err := cfg.db.GetChirps(nil)
	if err != nil {
		return nil, err
	}
	i := slices.IndexFunc(chirps, func(c entities.Chirp) bool {
		return c.Id == id
	})
	if i == -1 {
		return nil, errNotFound
	}
	return &chirps[i], nil
}

func (cfg *apiConfig) handlerListChirps(w http.ResponseWriter, req *http.Request) {
	userIdQuery := req.URL.Query().Get("author_id")
	var byUserId *int
	if userIdQuery != "" {
		v, err := strconv.Atoi(userIdQuery)
		if err != nil {
			respondWithError(w, 400, err.Error())
			return
		}
		byUserId = &v
	}

	chirps, err := cfg.db.GetChirps(byUserId)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	sortFn := func(i, j int) bool { return chirps[i].Id < chirps[j].Id }
	sortQuery := req.URL.Query().Get("sort")
	if sortQuery != "" && sortQuery != "asc" {
		if sortQuery == "desc" {
			sortFn = func(i, j int) bool { return chirps[i].Id > chirps[j].Id }
		} else {
			respondWithError(w, 400, "invalid sort query parameter")
			return
		}
	}
	sort.Slice(chirps, sortFn)

	respondWithJSON(w, 200, chirps)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, req *http.Request) {
	chirpId, err := strconv.Atoi(req.PathValue("chirpId"))
	if err != nil {
		respondWithError(w, 400, "invalid integer for chirp id")
		return
	}
	chirp, err := cfg.findChirpById(chirpId)
	if err != nil {
		if errors.Is(err, errNotFound) {
			respondWithError(w, 404, "chirp not found")
		} else {
			respondWithError(w, 500, err.Error())
		}
		return
	}
	respondWithJSON(w, 200, chirp)
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

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, req *http.Request) {
	userId, err := cfg.isAuthenticated(req)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}
	chirpId, err := strconv.Atoi(req.PathValue("chirpId"))
	if err != nil {
		respondWithError(w, 400, "invalid integer for chirp id")
		return
	}
	chirp, err := cfg.findChirpById(chirpId)
	if err != nil {
		if errors.Is(err, errNotFound) {
			respondWithError(w, 404, "chirp not found")
		} else {
			respondWithError(w, 500, err.Error())
		}
		return
	}
	if chirp.UserId != userId {
		respondWithError(w, 403, "forbidden")
		return
	}
	if err := cfg.db.DeleteChirp(chirp.Id); err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	respondWithJSON(w, 204, struct{}{})
}
