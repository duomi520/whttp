package whttp

import (
	"fmt"
	"net/http"
	"strings"
)

// ValidatorMiddleware 输入参数验证
func ValidatorMiddleware(a ...string) func(*HTTPContext) {
	var s0, s1 []string
	for _, v := range a {
		s := strings.Split(v, ":")
		if len(s) != 2 {
			panic(fmt.Sprintf("validate:bad describe %s", v))
		}
		s0 = append(s0, s[0])
		s1 = append(s1, s[1])
	}
	return func(c *HTTPContext) {
		for i := range s0 {
			v := c.Request.PathValue(s0[i])
			if len(v) > 0 {
				if err := c.route.validatorVar(v, s1[i]); err != nil {
					c.String(http.StatusBadRequest, fmt.Sprintf("validate %s %s faile: %s", s0[i], s1[i], err.Error()))
					return
				}
			}
			v = c.Request.FormValue(s0[i])
			if len(v) > 0 {
				if err := c.route.validatorVar(v, s1[i]); err != nil {
					c.String(http.StatusBadRequest, fmt.Sprintf("validate %s %s faile: %s", s0[i], s1[i], err.Error()))
					return
				}
			}
		}
		c.Next()
	}
}
