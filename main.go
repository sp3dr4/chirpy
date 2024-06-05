package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

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

func handlerChirpValidation(w http.ResponseWriter, req *http.Request) {
	status := 200

	type chirpReq struct {
		Body string `json:"body"`
	}
	type respErr struct {
		Error string `json:"error"`
	}
	type respOk struct {
		Valid bool `json:"valid"`
	}

	var dat []byte
	chirp := chirpReq{}
	if err := json.NewDecoder(req.Body).Decode(&chirp); err != nil {
		status = 400
		dat, _ = json.Marshal(respErr{Error: "Something went wrong"})
	} else {
		if len(chirp.Body) > 140 {
			status = 400
			dat, _ = json.Marshal(respErr{Error: "Chirp is too long"})
		} else {
			dat, _ = json.Marshal(respOk{Valid: true})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(dat)
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
