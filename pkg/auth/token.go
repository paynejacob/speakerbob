package auth

import (
	"github.com/google/uuid"
	"net/http"
	"strings"
	"time"
)

type TokenType int

const (
	Invalid TokenType = iota
	Session
	Bearer
	Websocket
)

//go:generate go run github.com/paynejacob/hotcereal providergen github.com/paynejacob/speakerbob/pkg/auth.Token
type Token struct {
	Id        string    `json:"id,omitempty" hotcereal:"key"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	Name      string    `json:"name"`

	Token     string    `json:"-" hotcereal:"lookup"`
	Type      TokenType `json:"-"`
	UserId    string    `json:"-"`
	ExpiresAt time.Time `json:"-"`
}

func NewToken() Token {
	return Token{
		Id:        strings.Replace(uuid.New().String(), "-", "", 4),
		Token:     strings.Replace(uuid.New().String(), "-", "", 4),
		CreatedAt: time.Now(),
	}
}

func (p *TokenProvider) FromRequest(r *http.Request) *Token {
	var t string
	var token *Token
	var expectedType TokenType
	var cookie *http.Cookie

	if cookie, _ = r.Cookie(cookieName); cookie != nil {
		t = cookie.Value
		expectedType = Session
	} else if t = strings.TrimPrefix(r.Header.Get(authorizationHeader), authorizationHeaderValuePrefix); t != "" {
		expectedType = Bearer
	} else if t = r.URL.Query().Get(wsTokenParameterName); t != "" {
		expectedType = Websocket
	}

	token = p.GetByToken(t)

	// if the token does not exist return nil
	if token == nil {
		return nil
	}

	// if the token is not the expected type do not return it
	if token.Type != expectedType {
		return nil
	}

	return token
}
