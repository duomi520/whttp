package whttp

import (
	"encoding/base64"
	"net/http"
	"strings"
)

func BasicAuthMiddleware(valid func(c *HTTPContext, username, password string) bool) func(*HTTPContext) {
	if valid == nil {
		panic("BasicAuthMiddleware 验证函数不为nil")
	}
	return func(c *HTTPContext) {
		auth := c.Request.Header.Get("Authorization")
		const prefix = "Basic "
		l := len(prefix)
		if len(auth) > len(prefix) && strings.EqualFold(auth[:l], prefix) {
			encoded := strings.TrimSpace(auth[l:])
			decoded, err := base64.StdEncoding.DecodeString(encoded)
			if err == nil {
				cred := string(decoded)
				name, pw, found := strings.Cut(cred, ":")
				if found && len(name) > 0 {
					ok := valid(c, name, pw)
					if ok {
						c.Next()
						return
					}
				}

			}
		}
		c.Writer.Header().Set("WWW-Authenticate", "Base realm=\"Restricted\"")
		c.String(http.StatusUnauthorized, "Unauthorized")
	}
}

//https://github.com/labstack/echo/blob/master/middleware/basic_auth.go
