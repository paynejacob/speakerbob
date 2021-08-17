// +build development

package server

import (
	"github.com/paynejacob/speakerbob/pkg/auth"
	"net/http"
	"net/url"
)

func init() {
	DefaultConfiguration.providers = append(DefaultConfiguration.providers, DevAuthProvider{})
	DefaultConfiguration.DataPath = ".speakerbob"
	DefaultConfiguration.Host = "127.0.0.1"
	DefaultConfiguration.Port = 8080
}

type DevAuthProvider struct{}

func (d DevAuthProvider) Name() string {
	return "debug"
}

func (d DevAuthProvider) VerifyCallback(r *http.Request) (principal auth.Principal, userEmail string, err error) {
	return auth.NewPrincipal(d.Name(), "1"), "u@d.co", nil
}

func (d DevAuthProvider) LoginRedirect(w http.ResponseWriter, r *http.Request, state string) {
	values := make(url.Values, 1)

	values.Add("state", state)

	http.Redirect(w, r, "/auth/callback/?"+values.Encode(), http.StatusFound)
}
