package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"

	"github.com/sp3dr4/chirpy/internal/database"
	"github.com/sp3dr4/chirpy/internal/entities"
)

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
	db             *database.DB
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

func (cfg *apiConfig) handlerListChirps(w http.ResponseWriter, req *http.Request) {
	chirps, err := cfg.db.GetChirps()
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	sort.Slice(chirps, func(i, j int) bool { return chirps[i].Id < chirps[j].Id })
	respondWithJSON(w, 200, chirps)
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, req *http.Request) {
	type chirpRequest struct {
		Body string `json:"body"`
	}
	chirpReq := chirpRequest{}
	if err := json.NewDecoder(req.Body).Decode(&chirpReq); err != nil {
		respondWithError(w, 400, "error decoding request body")
		return
	}
	cleaned, err := entities.ValidateChirp(chirpReq.Body)
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}
	chirp, err := cfg.db.CreateChirp(cleaned)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	respondWithJSON(w, 201, chirp)
}

func handlerHealth(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	io.WriteString(w, "OK")
}

func main() {
	db, err := database.NewDB("database.json")
	if err != nil {
		log.Fatalf("error with database initialization: %s", err)
	}
	cfg := apiConfig{fileserverHits: 0, db: db}
	mux := http.NewServeMux()
	fileSv := http.FileServer(http.Dir("."))
	mux.Handle("/app/*", http.StripPrefix("/app", cfg.middlewareMetricsInc(fileSv)))
	mux.HandleFunc("/api/reset", cfg.handlerResetMetrics)
	mux.HandleFunc("GET /admin/metrics", cfg.handlerGetMetrics)
	mux.HandleFunc("GET /api/healthz", handlerHealth)
	mux.HandleFunc("POST /api/chirps", cfg.handlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", cfg.handlerListChirps)
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	server.ListenAndServe()
}
