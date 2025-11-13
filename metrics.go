package main

import (
	"fmt"
	"net/http"
	"strconv"
)

func (cfg *apiConfig) Metrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	count := strconv.FormatUint(uint64(cfg.fileserverHits.Load()), 10)
	body := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %s times!</p>
  </body>
</html>`,
		count)
	_, err := w.Write([]byte(body))
	if err != nil {
		fmt.Printf("Error writing response: %v\n", err)
	}
}
