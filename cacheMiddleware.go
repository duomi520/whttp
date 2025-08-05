package whttp

import (
	"bytes"
	"net/http"
)

// Cache 缓存接口
type Cache interface {
	Del(key string)
	Set(key string, value []byte)
	HasGet(dst []byte, key string) ([]byte, bool)
}

func CacheMiddleware(cache Cache, header map[string]string) func(*HTTPContext) {
	return func(c *HTTPContext) {
		key := c.Request.URL.RequestURI()
		data, ok := cache.HasGet(nil, key)
		if ok {
			for k, v := range header {
				c.Writer.Header().Set(k, v)
			}
			c.write(http.StatusOK, data)
			return
		}
		f := func(b *bytes.Buffer) *bytes.Buffer {
			cache.Set(key, b.Bytes())
			return b
		}
		c.HookBeforWriteHeader = append(c.HookBeforWriteHeader, f)
		c.Next()
	}
}

// https://pkg.go.dev/github.com/VictoriaMetrics/fastcache
