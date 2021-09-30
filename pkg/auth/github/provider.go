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
	githubOrgsURL          = "https://api.github.com/user/orgs?per_page=100"
	githubEmailsURL        = "https://api.github.com/user/emails?per_page=100"
	githubScope            = "read:org,user:email"
)

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

	// https://docs.github.com/en/rest/reference/orgs#list-organizations-for-the-authenticated-user
	req, _ = http.NewRequest(http.MethodGet, githubOrgsURL, nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "token "+ghToken)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		logrus.Errorf("bad response for github orgs: [%d]", resp.StatusCode)
		err = errors.New("bad response for github orgs")
		return
	}

	// TODO: support paging here, users with >100 orgs will fail
	// https://docs.github.com/en/rest/reference/users#get-the-authenticated-user
	var orgList []struct {
		Login string `json:"login"`
	}

	err = json.NewDecoder(resp.Body).Decode(&orgList)
	if err != nil {
		return "", "", nil, err
	}

	for _, org := range orgList {
		orgs = append(orgs, org.Login)
	}

	return
}
