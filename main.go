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
	"regexp"
	"sort"
	"strings"

	stemmer "github.com/agonopol/go-stem"
)

type (
	// ShakespeareDataRec holds all lines of Shakespear's works
	ShakespeareDataRec struct {
		ShakespeareLine []ShakespeareLineRec `json:"data"`
	}
	// ShakespeareLineRec hold individual lines of Shakespear's works
	ShakespeareLineRec struct {
		Type         string   `json:"type,omitempty"`
		LineID       int      `json:"line_id,omitempty"`
		PlayName     string   `json:"play_name,omitempty"`
		SpeechNumber string   `json:"speech_number,omitempty"`
		LineNumber   string   `json:"line_number,omitempty"`
		Speaker      string   `json:"speaker,omitempty"`
		TextEntry    string   `json:"text_entry,omitempty"`
		WordsList    []string `json:"words_list,omitempty"`
	}
	// ResultRec hold the data from the database used to fill the template
	ResultRec struct {
		lineBefore string
		hit        string
		lineAfter  string
		playName   string
		lineNumber string
		speaker    string
		lineType   string
		score      int
	}
	// ResultsToFront used for JSON object to frontend
	ResultsToFront struct {
		HTML       string `json:"HTML"`
		NumResults int    `json:"numResults"`
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
	b, err := ioutil.ReadFile("shakeworks.json")
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

	// Init http server
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
		resultsToFront := data.Search(query[0])
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		err := enc.Encode(resultsToFront)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

// Search main search function
func (s *ShakespeareDataRec) Search(query string) (resultsToFront ResultsToFront) {

	var (
		hits    []ResultRec
		results string
	)

	queryWords := getWordsSlice(query)

	for i := range s.ShakespeareLine {
		score := matchScore(s.ShakespeareLine[i].WordsList, queryWords)
		if score > 0 {
			hit := ResultRec{
				lineBefore: s.ShakespeareLine[i-1].TextEntry,
				hit:        makeBold(s.ShakespeareLine[i].TextEntry, query),
				lineAfter:  s.ShakespeareLine[i+1].TextEntry,
				playName:   s.ShakespeareLine[i].PlayName,
				lineNumber: s.ShakespeareLine[i].LineNumber,
				speaker:    s.ShakespeareLine[i].Speaker,
				lineType:   s.ShakespeareLine[i].Type,
				score:      score,
			}
			hits = append(hits, hit)
		}
	}

	sortResults(hits)

	for _, v := range hits {
		results += v.FormatHTML(v)
	}
	resultsToFront = ResultsToFront{
		HTML:       results,
		NumResults: len(hits),
	}
	return resultsToFront
}

func getWordsSlice(s string) (wordSlice []string) {

	// Make a Regex to say we only want letters, numbers, and spaces
	reg, err := regexp.Compile("[^a-zA-Z0-9\\s]+")
	if err != nil {
		log.Println(err)
	}
	processedString := reg.ReplaceAllString(s, "")

	words := strings.Split(processedString, " ")

	for _, v := range words {
		wordSlice = append(wordSlice, string(stemmer.Stem([]byte(v))))
	}
	return wordSlice
}

func matchScore(data []string, query []string) (score int) {
	for _, d := range data {
		for _, q := range query {
			if d == q {
				score++
			}
		}
	}
	return score
}

func makeBold(line string, query string) string {
	slice := strings.Split(line, " ")
	for i := range slice {
		if strings.Contains(
			strings.ToLower(slice[i]),
			strings.ToLower(query),
		) {
			slice[i] = fmt.Sprintf("<b>%v</b>", slice[i])
		}
	}

	return strings.Join(slice, " ")
}

func sortResults(hits []ResultRec) {
	var less func(i, j int) bool

	less = func(i, j int) bool {
		return hits[i].score > hits[j].score
	}

	sort.Slice(hits, less)
}

// FormatHTML formats the data in the needed HTML
func (r *ResultRec) FormatHTML(hit ResultRec) string {
	text := fmt.Sprintf("%v %v %v", r.lineBefore, r.hit, r.lineAfter)
	formatedResult := fmt.Sprintf(`<figure class="result">
        <div class="result__content">
            <div class="result__title">
                <h2 class="result__heading">%v %v</h2>
                <div class="result__tag result__tag--1">#%v</div>
				<div class="result__tag result__tag--2">#%v</div>
            </div>
            <p class="result__description">%v</p>
        </div>
        <div class="result__work">
        	RANK %v
        </div>
	</figure>`, r.playName, r.lineNumber, r.lineType, r.speaker, text, r.score)

	return formatedResult

}

// SearchSimple simple string match function
func (s *ShakespeareDataRec) SearchSimple(query string) string {

	var results string

	for i := range s.ShakespeareLine {
		if strings.Contains(
			strings.ToLower(s.ShakespeareLine[i].TextEntry),
			strings.ToLower(query),
		) {
			hit := ResultRec{
				lineBefore: s.ShakespeareLine[i-1].TextEntry,
				hit:        strings.Replace(s.ShakespeareLine[i].TextEntry, query, fmt.Sprintf("<b>%v</b>", query), -1),
				lineAfter:  s.ShakespeareLine[i+1].TextEntry,
				playName:   s.ShakespeareLine[i].PlayName,
				lineNumber: s.ShakespeareLine[i].LineNumber,
				speaker:    s.ShakespeareLine[i].Speaker,
				lineType:   s.ShakespeareLine[i].Type,
			}
			results += hit.FormatHTML(hit)
		}
	}

	return results
}
