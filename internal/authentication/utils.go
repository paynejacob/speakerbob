package authentication

import (
	"fmt"
	"net/http"
)

func getCookieKey(addr string, cookie *http.Cookie) string {
	return fmt.Sprintf("%s::%s", addr, cookie.Value)
}
