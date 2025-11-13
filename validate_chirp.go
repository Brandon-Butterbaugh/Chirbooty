package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func validate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		// an error will be thrown if the JSON is invalid or has the wrong types
		// any missing fields will simply have their values in the struct set to their zero value
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	type returnJSON struct {
		Error   string `json:"error"`
		Cleaned string `json:"cleaned_body"`
	}
	var code int
	var respBody returnJSON

	if len(params.Body) > 140 {
		respBody.Error = "Chirp is too long"
		code = 400
	} else {
		respBody.Cleaned = cleanProfanity(params.Body)
		code = 200
	}

	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func cleanProfanity(str string) string {
	temp := str
	temp = strings.ToLower(temp)
	split := strings.Split(temp, " ")
	strSplit := strings.Split(str, " ")

	for i, word := range split {
		if word == "kerfuffle" || word == "sharbert" || word == "fornax" {
			strSplit[i] = "****"
		}
	}
	return strings.Join(strSplit, " ")
}
