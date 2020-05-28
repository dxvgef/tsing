package middleware

import (
	"net/http"
	"strings"

	"github.com/dxvgef/tsing"
)

var xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
var xRealIP = http.CanonicalHeaderKey("X-Real-IP")

func RealIP(c *tsing.Context) error {
	if rip := realIP(c.Request); rip != "" {
		c.Request.RemoteAddr = rip
	}
	return nil
}

func realIP(r *http.Request) string {
	var ip string

	if xrip := r.Header.Get(xRealIP); xrip != "" {
		ip = xrip
	} else if xff := r.Header.Get(xForwardedFor); xff != "" {
		i := strings.Index(xff, ", ")
		if i == -1 {
			i = len(xff)
		}
		ip = xff[:i]
	}

	return ip
}
