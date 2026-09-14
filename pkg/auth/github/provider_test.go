package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetGithubOrgs_SinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "token test-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"login":"org-a"},{"login":"org-b"}]`))
	}))
	defer server.Close()

	originalURL := githubOrgsURL
	githubOrgsURL = server.URL
	defer func() { githubOrgsURL = originalURL }()

	orgs, err := getGithubOrgs("test-token")

	require.NoError(t, err)
	assert.Equal(t, []string{"org-a", "org-b"}, orgs)
}

func TestGetGithubOrgs_FollowsPagination(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Query().Get("page") {
		case "":
			w.Header().Set("Link", fmt.Sprintf(`<%s?page=2>; rel="next", <%s?page=3>; rel="last"`, server.URL, server.URL))
			_, _ = w.Write([]byte(`[{"login":"org-a"}]`))
		case "2":
			w.Header().Set("Link", fmt.Sprintf(`<%s?page=1>; rel="prev", <%s?page=3>; rel="next", <%s?page=3>; rel="last"`, server.URL, server.URL, server.URL))
			_, _ = w.Write([]byte(`[{"login":"org-b"}]`))
		case "3":
			// last page: no rel="next" link
			w.Header().Set("Link", fmt.Sprintf(`<%s?page=2>; rel="prev", <%s?page=1>; rel="first"`, server.URL, server.URL))
			_, _ = w.Write([]byte(`[{"login":"org-c"}]`))
		default:
			t.Fatalf("unexpected page requested: %q", r.URL.Query().Get("page"))
		}
	}))
	defer server.Close()

	originalURL := githubOrgsURL
	githubOrgsURL = server.URL
	defer func() { githubOrgsURL = originalURL }()

	orgs, err := getGithubOrgs("test-token")

	require.NoError(t, err)
	assert.Equal(t, []string{"org-a", "org-b", "org-c"}, orgs)
}

func TestGetGithubOrgs_BadResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	originalURL := githubOrgsURL
	githubOrgsURL = server.URL
	defer func() { githubOrgsURL = originalURL }()

	orgs, err := getGithubOrgs("test-token")

	require.Error(t, err)
	assert.Nil(t, orgs)
}
