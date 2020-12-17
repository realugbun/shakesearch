package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type (
	// ShakespeareDataRec holds all lines of Shakespear's works
	ShakespeareDataRec struct {
		ShakespeareLine []ShakespeareLineRec `json:"data"`
	}
	// ShakespeareLineRec hold individual lines of Shakespear's works
	ShakespeareLineRec struct {
		Type         string      `json:"type,omitempty"`
		LineID       int         `json:"line_id,omitempty"`
		PlayName     string      `json:"play_name,omitempty"`
		SpeechNumber interface{} `json:"speech_number,omitempty"`
		LineNumber   string      `json:"line_number,omitempty"`
		Speaker      string      `json:"speaker,omitempty"`
		TextEntry    string      `json:"text_entry,omitempty"`
	}
)

var (
	data ShakespeareDataRec
	err  error
)

func main() {

	// INITIALIZATION
	// Init logs
	lf, err := os.OpenFile("shakespeare.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		err = errors.New("Failed to open/create log file")
		log.Println(err)
	}
	log.SetOutput(lf)
	defer lf.Close()

	// Get input from Database
	b, err := ioutil.ReadFile("data.json")
	if err != nil {
		err = errors.New("Failed to open database file")
		log.Fatal(err)
	}

	err = json.Unmarshal([]byte(b), &data)
	if err != nil {
		err = errors.New("Failed to unmarshal database")
		log.Fatal(err)
	}

	log.Println("Database loaded succussfully")

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(data))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("Listening on port %s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handleSearch(data ShakespeareDataRec) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}
		results := data.Search(query[0])
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		err := enc.Encode(results)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

// Search method for finding a string within the data
func (s *ShakespeareDataRec) Search(query string) []string {

	var results []string

	for i := range s.ShakespeareLine {
		if strings.Contains(s.ShakespeareLine[i].TextEntry, query) {
			qureyString := strings.Replace(s.ShakespeareLine[i].TextEntry, query, fmt.Sprintf("<b>%v</b>", query), -1)
			results = append(results, fmt.Sprintf("\"%v %v %v\" (%v %v)", s.ShakespeareLine[i-1].TextEntry, qureyString, s.ShakespeareLine[i+1].TextEntry, s.ShakespeareLine[i].PlayName, s.ShakespeareLine[i].LineNumber))
		}
	}
	return results
}
