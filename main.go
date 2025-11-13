package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, resp *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, resp)
	})

}

func main() {
	const filepathRoot = "."
	const port = "8080"
	mux := http.NewServeMux()
	serv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	cfg := &apiConfig{
		fileserverHits: atomic.Int32{},
	}
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))

	mux.Handle("/app/", cfg.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /healthz", health)
	mux.HandleFunc("GET /metrics", cfg.Metrics)
	mux.HandleFunc("POST /reset", cfg.Reset)

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(serv.ListenAndServe())

}
