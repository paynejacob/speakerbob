package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"strconv"
)

type DisplayResult struct {
	Type   string      `json:"type"`
	Result interface{} `json:"result"`
}

type SearchService struct {
	backend SearchBackend
}

func NewSearchService(backendURL string) *SearchService {
	var backend SearchBackend
	parsedUrl, err := url.Parse(backendURL)
	if err != nil {
		panic("invalid backend url")
	}

	switch parsedUrl.Scheme {
	case "memory":
		backend = NewSearchMemoryBackend()
	default:
		panic(fmt.Sprintf("\"%s\" is not a valid search backend url", parsedUrl.Scheme))
	}

	return &SearchService{backend}
}

func (s *SearchService) RegisterRoutes(parent *mux.Router, prefix string) *mux.Router {
	router := parent.PathPrefix(prefix).Subrouter()

	router.HandleFunc("/", s.Search).Methods("GET")

	return router
}

func (s *SearchService) Search(w http.ResponseWriter, r *http.Request) {
	count := 100
	displayResults := make([]DisplayResult, 0)
	var query string

	if queryString, ok := r.URL.Query()["query"]; ok {
		query = queryString[0]
	}

	if countStr, ok := r.URL.Query()["count"]; ok {
		count, _ = strconv.Atoi(countStr[0])
	}

	if query == "" {
		_ = json.NewEncoder(w).Encode([]string{})
		return
	}

	results, err := s.backend.Search(query, count)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, result := range results {
		displayResults = append(displayResults, DisplayResult{result.Type(), result.Object()})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(displayResults)
}

func (s *SearchService) UpdateResult(result Result) error {
	return s.backend.UpdateResult(result)
}
