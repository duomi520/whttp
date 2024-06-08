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
		basic := "Basic"
		l := len(basic)
		if len(auth) > l+1 && strings.EqualFold(auth[:l], basic) {
			b, err := base64.StdEncoding.DecodeString(auth[l+1:])
			if err != nil {
				c.String(http.StatusBadRequest, err.Error())
			}
			cred := string(b)
			for i := 0; i < len(cred); i++ {
				if cred[i] == ':' {
					ok := valid(c, cred[:i], cred[i+1:])
					if ok {
						c.Next()
						return
					}
					break
				}
			}
		}
		c.Writer.Header().Set("WWW-Authenticate", "Base realm=\"Restricted\" charset=\"UTF-8\"")
		c.String(http.StatusUnauthorized, "")
	}
}

//https://github.com/labstack/echo/blob/master/middleware/basic_auth.go
