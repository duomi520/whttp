package whttp

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

func checkCacheControl(req *http.Request) bool {
	cc := req.Header.Get("Cache-Control")
	if len(cc) < 7 {
		return false
	}
	// 统一转为小写，避免大小写敏感问题（HTTP头部值大小写不敏感）
	ccLower := strings.ToLower(cc)
	if strings.Contains(ccLower, "no-store") {
		return true
	}
	if strings.Contains(ccLower, "no-cache") {
		return true
	}
	if strings.Contains(ccLower, "must-revalidate") {
		return true
	}
	if strings.Contains(ccLower, "max-age=0") {
		return true
	}
	return false
}
func ETagMiddleware(etag *sync.Map) func(*HTTPContext) {
	return func(c *HTTPContext) {
		f := func(b *bytes.Buffer) *bytes.Buffer {
			if c.status == http.StatusOK {
				ETag := fmt.Sprintf("%x", (md5.Sum(b.Bytes())))
				etag.Store(c.Request.URL.RequestURI(), ETag)
				c.Writer.Header().Set("ETag", ETag)
			}
			return b
		}
		c.HookBeforWriteHeader = append(c.HookBeforWriteHeader, f)
		if !checkCacheControl(c.Request) {
			ETag, ok := etag.Load(c.Request.URL.RequestURI())
			if ok && strings.Contains(c.Request.Header.Get("If-None-Match"), ETag.(string)) {
				c.status = http.StatusNotModified
				c.Writer.WriteHeader(http.StatusNotModified)
				return
			}
		}
		c.Next()
	}
}
