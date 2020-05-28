package middleware

import (
	"fmt"
	"net/http"

	"github.com/dxvgef/tsing"
)

// BasicAuth implements a simple middleware handler for adding basic http auth to a route.
func BasicAuth(realm string, creds map[string]string) tsing.Handler {
	return func(c *tsing.Context) error {
		user, pass, ok := c.Request.BasicAuth()
		if !ok {
			basicAuthFailed(c.ResponseWriter, realm)
			c.Abort()
		}
		credPass, credUserOk := creds[user]
		if !credUserOk || pass != credPass {
			basicAuthFailed(c.ResponseWriter, realm)
			c.Abort()
		}
		return nil
	}
}

func basicAuthFailed(w http.ResponseWriter, realm string) {
	w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
	w.WriteHeader(http.StatusUnauthorized)
}
