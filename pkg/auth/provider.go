package auth

import (
	"net/http"
)

type ProviderError struct {
	Reason string
}

func (e ProviderError) Error() string {
	return e.Reason
}

type AccessDenied struct {
	ProviderError
}

type Provider interface {
	Name() string
	VerifyCallback(r *http.Request) (principal Principal, userEmail string, err error)
	LoginRedirect(w http.ResponseWriter, r *http.Request, state string)
}
