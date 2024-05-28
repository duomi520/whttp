package whttp

import (
	"net/http"

	"github.com/duomi520/utils"
)

// LimitMiddleware 令牌桶限流器
func LimitMiddleware(tl *utils.TokenBucketLimiter) func(*HTTPContext) {
	return func(c *HTTPContext) {
		if err := tl.Take(1); err == nil {
			c.Next()
		} else {
			c.String(http.StatusTooManyRequests, err.Error())
		}
	}
}

// https://cloud.tencent.com/developer/article/1685510
