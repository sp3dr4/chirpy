package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"

	"github.com/sp3dr4/chirpy/internal/entities"
)

type Payload struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

func (cfg *apiConfig) userUpgradedCallback(w http.ResponseWriter, payload Payload) {
	type UserUpgradedData struct {
		UserID int `json:"user_id"`
	}
	var data UserUpgradedData
	if err := json.Unmarshal(payload.Data, &data); err != nil {
		respondWithError(w, 400, err.Error())
		return
	}

	users, err := cfg.db.GetUsers()
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	i := slices.IndexFunc(users, func(c entities.User) bool {
		return c.Id == data.UserID
	})
	if i == -1 {
		respondWithError(w, 404, "user not found")
		return
	}
	user := users[i]
	user.IsChirpyRed = true
	_, err = cfg.db.UpdateUser(&user)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	respondWithJSON(w, 204, struct{}{})
}

func (cfg *apiConfig) userDowngradedCallback(w http.ResponseWriter, payload Payload) {
	type UserDowngradedData struct {
		UserID int  `json:"user_id"`
		Pause  bool `json:"pause"`
	}
	var data UserDowngradedData
	if err := json.Unmarshal(payload.Data, &data); err != nil {
		respondWithError(w, 400, err.Error())
		return
	}
}

func (cfg *apiConfig) handlerWebhookPolka(w http.ResponseWriter, r *http.Request) {
	apiKey, found := strings.CutPrefix(r.Header.Get("Authorization"), "ApiKey ")
	if !found {
		respondWithError(w, 401, "no authorization header")
		return
	}
	if apiKey != cfg.polkaApiKey {
		respondWithError(w, 401, "no authorization header")
		return
	}

	var polkaCallbacksMap = map[string]func(http.ResponseWriter, Payload){
		"user.upgraded":   cfg.userUpgradedCallback,
		"user.downgraded": cfg.userDowngradedCallback,
	}

	var payload Payload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondWithError(w, 400, "error decoding request body")
		return
	}

	callback, ok := polkaCallbacksMap[payload.Event]
	if !ok {
		respondWithJSON(w, 204, struct{}{})
		return
	}
	callback(w, payload)
}
