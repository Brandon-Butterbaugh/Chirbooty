package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/Brandon-Butterbaugh/Chirbooty.git/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	database       *database.Queries
	platform       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, resp *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, resp)
	})

}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("Error writing response: %v\n", err)
	}
	dbQueries := database.New(db)

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}

	const filepathRoot = "."
	const port = "8080"
	mux := http.NewServeMux()
	serv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	cfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		database:       dbQueries,
		platform:       platform,
	}
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))

	mux.Handle("/app/", cfg.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /api/healthz", health)
	mux.HandleFunc("GET /admin/metrics", cfg.Metrics)
	mux.HandleFunc("POST /admin/reset", cfg.Reset)
	mux.HandleFunc("POST /api/users", cfg.newUser)
	mux.HandleFunc("POST /api/chirps", cfg.newChirp)
	mux.HandleFunc("GET /api/chirps", cfg.getChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.getChirp)
	mux.HandleFunc("POST /api/login", cfg.login)

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(serv.ListenAndServe())

}
