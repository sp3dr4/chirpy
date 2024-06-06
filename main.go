package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
)

var profanities []string = []string{"kerfuffle", "sharbert", "fornax"}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type respErr struct {
		Error string `json:"error"`
	}
	dat, _ := json.Marshal(respErr{Error: msg})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, 500, "error encoding response")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

type apiConfig struct {
	fileserverHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits += 1
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerGetMetrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(200)
	fmt.Fprintf(w, "<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits)
}

func (cfg *apiConfig) handlerResetMetrics(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits = 0
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
}

func validateChirp(text string) (string, error) {
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

func handlerChirpValidation(w http.ResponseWriter, req *http.Request) {
	type chirpReq struct {
		Body string `json:"body"`
	}
	type respOk struct {
		Cleaned string `json:"cleaned_body"`
	}

	chirp := chirpReq{}
	if err := json.NewDecoder(req.Body).Decode(&chirp); err != nil {
		respondWithError(w, 400, "error decoding request body")
		return
	}
	cleaned, err := validateChirp(chirp.Body)
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}
	respondWithJSON(w, 200, respOk{Cleaned: cleaned})
}

func handlerHealth(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	io.WriteString(w, "OK")
}

func main() {
	cfg := apiConfig{}
	mux := http.NewServeMux()
	fileSv := http.FileServer(http.Dir("."))
	mux.Handle("/app/*", http.StripPrefix("/app", cfg.middlewareMetricsInc(fileSv)))
	mux.HandleFunc("GET /api/healthz", handlerHealth)
	mux.HandleFunc("GET /admin/metrics", cfg.handlerGetMetrics)
	mux.HandleFunc("/api/reset", cfg.handlerResetMetrics)
	mux.HandleFunc("POST /api/validate_chirp", handlerChirpValidation)
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	server.ListenAndServe()
}
