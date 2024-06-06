package entities

import (
	"errors"
	"slices"
	"strings"
)

var profanities []string = []string{"kerfuffle", "sharbert", "fornax"}

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

func ValidateChirp(text string) (string, error) {
	if len(text) > 140 {
		return "", errors.New("chirp is too long")
	}
	words := strings.Split(text, " ")
	for i, word := range words {
		if slices.Contains(profanities, strings.ToLower(word)) {
			words[i] = "****"
		}
	}
	return strings.Join(words, " "), nil
}
