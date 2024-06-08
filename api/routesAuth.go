package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/sp3dr4/chirpy/internal/entities"
	"golang.org/x/crypto/bcrypt"
)

func (cfg *apiConfig) isAuthenticated(r *http.Request) (int, error) {
	tokenStr, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !found {
		return 0, errors.New("no authorization header")
	}

	userId, err := getUserIdFromJwt(tokenStr, cfg.jwtSecret)
	if err != nil {
		return 0, errors.New("unauthorized")
	}
	return userId, nil
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, req *http.Request) {
	type loginResponse struct {
		userResponse
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	userReq := userRequest{}
	if err := json.NewDecoder(req.Body).Decode(&userReq); err != nil {
		respondWithError(w, 400, "error decoding request body")
		return
	}

	users, err := cfg.db.GetUsers()
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	i := slices.IndexFunc(users, func(c entities.User) bool {
		return c.Email == strings.ToLower(userReq.Email)
	})
	if i == -1 {
		respondWithError(w, 404, "user not found")
		return
	}
	user := users[i]
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userReq.Password)); err != nil {
		respondWithError(w, 401, "unauthorized")
		return
	}

	signedToken, err := createJwt(user.Id, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	refreshStr, err := buildRandomToken()
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	refreshToken, err := cfg.db.SaveRefreshToken(user.Id, refreshStr, buildExpiration(defaultRefreshExpirationSeconds))
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(
		w,
		200,
		loginResponse{
			userResponse: userResponse{Id: user.Id, Email: user.Email, IsChirpyRed: user.IsChirpyRed},
			Token:        signedToken,
			RefreshToken: refreshToken.Token,
		},
	)
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, req *http.Request) {
	refreshStr, found := strings.CutPrefix(req.Header.Get("Authorization"), "Bearer ")
	if !found {
		respondWithError(w, 401, "no authorization header")
		return
	}

	refreshObj, err := cfg.db.GetRefreshToken(refreshStr)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}
	if refreshObj.ExpiresAt.Before(time.Now()) {
		respondWithError(w, 401, "token expired")
		return
	}

	signedToken, err := createJwt(refreshObj.UserId, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(
		w,
		200,
		struct {
			Token string `json:"token"`
		}{Token: signedToken},
	)
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, req *http.Request) {
	refreshStr, found := strings.CutPrefix(req.Header.Get("Authorization"), "Bearer ")
	if !found {
		respondWithError(w, 401, "no authorization header")
	}

	if err := cfg.db.DeleteRefreshToken(refreshStr); err != nil {
		respondWithError(w, 401, err.Error())
		return
	}
	respondWithJSON(w, 204, struct{}{})
}
