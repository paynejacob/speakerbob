package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/paynejacob/speakerbob/pkg/auth"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	githubAuthorizationURL = "https://github.com/login/oauth/authorize"
	githubAccessTokenURL   = "https://github.com/login/oauth/access_token"
	githubUserURL          = "https://api.github.com/user"
	githubEmailsURL        = "https://api.github.com/user/emails?per_page=100"
	githubScope            = "read:org,user:email"
)

// githubOrgsURL is a var (not const) so tests can point it at a mock server.
var githubOrgsURL = "https://api.github.com/user/orgs?per_page=100"

// https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps

type Provider struct {
	Enabled      bool   `yaml:"enabled"`
	ClientId     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`

	OrganizationPermissionMap map[string]bool `yaml:"organization_permission_map"`
	EmailPermissionMap        map[string]bool `yaml:"email_permission_map"`
}

func (g Provider) Name() string {
	return "github"
}

func (g Provider) VerifyCallback(r *http.Request) (principal auth.Principal, userEmail string, err error) {
	var userId string
	var orgs []string
	var allowed bool

	logrus.Debug("[github] exchanging callback code for token")
	token, err := getGithubToken(g.ClientId, g.ClientSecret, r.URL.Query().Get("code"))
	if err != nil {
		logrus.Errorf("error getting github token: %v", err)
		return
	}

	logrus.Debug("[github] requesting user info for callback")
	userId, userEmail, orgs, err = getGithubUserInfo(token)
	if err != nil {
		logrus.Errorf("error getting github user info: %v", err)
		return
	}

	principal = auth.NewPrincipal(g.Name(), userId)

	logrus.Debugf("[github] checking github orgs for user: %s", userId)
	for _, org := range orgs {
		if g.OrganizationPermissionMap[org] {
			allowed = true
			break
		}
	}
	logrus.Debugf("[github] user allowed based on org access? %s => %v", userId, allowed)

	if !allowed {
		allowed = g.EmailPermissionMap[userEmail]
	}
	logrus.Debugf("[github] user allowed based on email access? %s [%s] => %v", userId, userEmail, allowed)

	if !allowed {
		err = auth.AccessDenied{}
	}

	return
}

func (g Provider) LoginRedirect(w http.ResponseWriter, r *http.Request, state string) {
	values := make(url.Values, 4)

	values.Add("client_id", g.ClientId)
	values.Add("scope", githubScope)
	values.Add("state", state)

	http.Redirect(w, r, githubAuthorizationURL+fmt.Sprintf("?%s", values.Encode()), http.StatusFound)
}

func getGithubToken(clientId, clientSecret, code string) (string, error) {
	var err error
	var resp *http.Response
	var req *http.Request

	values := make(url.Values, 4)

	values.Add("client_id", clientId)
	values.Add("client_secret", clientSecret)
	values.Add("code", code)

	req, _ = http.NewRequest(http.MethodPost, githubAccessTokenURL, bytes.NewReader([]byte(values.Encode())))
	req.Header.Set("Accept", "application/json")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		logrus.Errorf("Received invalid respnonse status getting github token: [%d]", resp.StatusCode)
		return "", errors.New("bad token response")
	}

	var body struct {
		AccessToken string `json:"access_token"`
	}

	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return "", err
	}

	return body.AccessToken, nil
}

func getGithubUserInfo(ghToken string) (userId string, userEmail string, orgs []string, err error) {
	var resp *http.Response
	var req *http.Request

	req, _ = http.NewRequest(http.MethodGet, githubUserURL, nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "token "+ghToken)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		logrus.Errorf("bad response for github user: [%d]", resp.StatusCode)
		err = errors.New("bad response for github user")
		return
	}

	// https://docs.github.com/en/rest/reference/users#get-the-authenticated-user
	var user struct {
		Id int `json:"id"`
	}

	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return
	}

	userId = strconv.Itoa(user.Id)

	req, _ = http.NewRequest(http.MethodGet, githubEmailsURL, nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "token "+ghToken)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		logrus.Errorf("bad response for github emails: [%d]", resp.StatusCode)
		err = errors.New("bad response for github emails")
		return
	}

	// https://docs.github.com/en/rest/reference/users?query=email#list-email-addresses-for-the-authenticated-user
	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}

	err = json.NewDecoder(resp.Body).Decode(&emails)
	if err != nil {
		return
	}

	for _, email := range emails {
		if email.Primary {
			userEmail = strings.ToLower(email.Email)
			break
		}
	}

	orgs, err = getGithubOrgs(ghToken)
	if err != nil {
		return "", "", nil, err
	}

	return
}

// getGithubOrgs fetches every org for the authenticated user, following the
// Link "next" header across pages instead of assuming the first page (up to
// per_page=100 orgs) is the complete list.
// https://docs.github.com/en/rest/reference/orgs#list-organizations-for-the-authenticated-user
func getGithubOrgs(ghToken string) (orgs []string, err error) {
	nextURL := githubOrgsURL

	for nextURL != "" {
		req, _ := http.NewRequest(http.MethodGet, nextURL, nil)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "token "+ghToken)

		resp, doErr := http.DefaultClient.Do(req)
		if doErr != nil {
			return nil, doErr
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			logrus.Errorf("bad response for github orgs: [%d]", resp.StatusCode)
			return nil, errors.New("bad response for github orgs")
		}

		var orgList []struct {
			Login string `json:"login"`
		}

		decodeErr := json.NewDecoder(resp.Body).Decode(&orgList)
		nextURL = parseNextLink(resp.Header.Get("Link"))
		resp.Body.Close()
		if decodeErr != nil {
			return nil, decodeErr
		}

		for _, org := range orgList {
			orgs = append(orgs, org.Login)
		}
	}

	return orgs, nil
}

// parseNextLink extracts the rel="next" URL from a GitHub API Link header,
// e.g. `<https://api.github.com/user/orgs?page=2>; rel="next", <...>; rel="last"`.
// Returns "" once there is no next page.
func parseNextLink(linkHeader string) string {
	for _, part := range strings.Split(linkHeader, ",") {
		segments := strings.Split(part, ";")
		if len(segments) < 2 {
			continue
		}

		url := strings.Trim(strings.TrimSpace(segments[0]), "<>")

		for _, seg := range segments[1:] {
			if strings.TrimSpace(seg) == `rel="next"` {
				return url
			}
		}
	}

	return ""
}
