package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/sp3dr4/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits int
	jwtSecret      string
	db             *database.DB
}

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

func handlerHealth(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	io.WriteString(w, "OK")
}

func main() {
	godotenv.Load()
	jwtSecret := os.Getenv("JWT_SECRET")

	isDebug := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	db, err := database.NewDB("database.json", *isDebug)
	if err != nil {
		log.Fatalf("error with database initialization: %s", err)
	}
	cfg := apiConfig{
		fileserverHits: 0,
		jwtSecret:      jwtSecret,
		db:             db,
	}
	mux := http.NewServeMux()
	fileSv := http.FileServer(http.Dir("."))
	mux.Handle("/app/*", http.StripPrefix("/app", cfg.middlewareMetricsInc(fileSv)))
	mux.HandleFunc("/api/reset", cfg.handlerResetMetrics)
	mux.HandleFunc("GET /admin/metrics", cfg.handlerGetMetrics)
	mux.HandleFunc("GET /api/healthz", handlerHealth)
	mux.HandleFunc("POST /api/chirps", cfg.handlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", cfg.handlerListChirps)
	mux.HandleFunc("GET /api/chirps/{chirpId}", cfg.handlerGetChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpId}", cfg.handlerDeleteChirp)
	mux.HandleFunc("POST /api/users", cfg.handlerCreateUser)
	mux.HandleFunc("PUT /api/users", cfg.handlerUpdateUser)
	mux.HandleFunc("POST /api/login", cfg.handlerLogin)
	mux.HandleFunc("POST /api/refresh", cfg.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", cfg.handlerRevoke)
	mux.HandleFunc("POST /api/polka/webhooks", cfg.handlerWebhookPolka)
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	log.Fatal(server.ListenAndServe())
}
