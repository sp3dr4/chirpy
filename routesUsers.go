package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/sp3dr4/chirpy/internal/database"
	"github.com/sp3dr4/chirpy/internal/entities"
	"golang.org/x/crypto/bcrypt"
)

type userRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, req *http.Request) {
	userReq := userRequest{}
	if err := json.NewDecoder(req.Body).Decode(&userReq); err != nil {
		respondWithError(w, 400, "error decoding request body")
		return
	}
	paswHash, err := bcrypt.GenerateFromPassword([]byte(userReq.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, 500, "something went wrong")
		return
	}
	user, err := cfg.db.CreateUser(strings.ToLower(userReq.Email), string(paswHash))
	if err != nil {
		code := 500
		if errors.Is(err, database.ErrDuplicateUser) {
			code = 400
		}
		respondWithError(w, code, err.Error())
		return
	}
	respondWithJSON(w, 201, userResponse{Id: user.Id, Email: user.Email})
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, req *http.Request) {
	tokenStr, found := strings.CutPrefix(req.Header.Get("Authorization"), "Bearer ")
	if !found {
		respondWithError(w, 401, "no authorization header")
	}

	userId, err := getUserIdFromJwt(tokenStr, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, 401, "unauthorized")
		return
	}

	userReq := userRequest{}
	if err := json.NewDecoder(req.Body).Decode(&userReq); err != nil {
		log.Printf("json.NewDecoder err: %v\n", err)
		respondWithError(w, 400, "error decoding request body")
		return
	}

	users, err := cfg.db.GetUsers()
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	i := slices.IndexFunc(users, func(c entities.User) bool {
		return c.Id == userId
	})
	if i == -1 {
		respondWithError(w, 404, "user not found")
		return
	}
	user := users[i]
	paswHash, err := bcrypt.GenerateFromPassword([]byte(userReq.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, 500, "something went wrong")
		return
	}
	if user.Email != strings.ToLower(userReq.Email) || user.Password != string(paswHash) {
		user.Email = strings.ToLower(userReq.Email)
		user.Password = string(paswHash)
		updated, err := cfg.db.UpdateUser(&user)
		if err != nil {
			respondWithError(w, 500, "something went wrong")
			return
		}
		user = *updated
	}

	respondWithJSON(w, 200, userResponse{Id: user.Id, Email: user.Email})
}
