package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type Service struct {
	backend Backend
}

func (s *Service) Search(w http.ResponseWriter, r *http.Request) {
	count := 100
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

	results := s.backend.Search(query, count)

	_, _ = w.Write([]byte("["))
	for result := range results {
		_, _ = w.Write([]byte(fmt.Sprintf("%s,", result)))
	}
	_, _ = w.Write([]byte("]"))
}
