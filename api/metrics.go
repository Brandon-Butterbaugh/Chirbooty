package main

import (
	"fmt"
	"net/http"
	"strconv"
)

func (cfg *apiConfig) Metrics(w http.ResponseWriter, req *http.Request) {
	count := strconv.FormatUint(uint64(cfg.fileserverHits.Load()), 10)
	body := fmt.Sprintf("Hits: %s\n", count)
	_, err := w.Write([]byte(body))
	if err != nil {
		fmt.Printf("Error writing response: %v\n", err)
	}
}
