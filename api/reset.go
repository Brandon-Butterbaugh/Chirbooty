package api

import (
	"net/http"
)

func (cfg *apiConfig) Reset(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
}
