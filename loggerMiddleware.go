package whttp

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

var formatLogger string = "| %-13v | %-15s | %3d | %-7s | %-40s | %7d bytes "

// LoggerMiddleware 日志
func LoggerMiddleware() func(*HTTPContext) {
	return func(c *HTTPContext) {
		var n int
		f := func(b *bytes.Buffer) *bytes.Buffer {
			n = b.Len()
			return b
		}
		c.HookBeforWriteHeader = append(c.HookBeforWriteHeader, f)
		startTime := time.Now()
		c.Next()
		latency := time.Since(startTime)
		// 智能截断过长的延迟显示
		if latency > time.Minute {
			latency = latency.Truncate(time.Millisecond)
		}
		switch {
		case c.status >= http.StatusInternalServerError:
			slog.Error(fmt.Sprintf(formatLogger, latency, c.Request.RemoteAddr, c.status, c.Request.Method, c.Request.URL, n))
		case c.status >= http.StatusBadRequest:
			slog.Warn(fmt.Sprintf(formatLogger, latency, c.Request.RemoteAddr, c.status, c.Request.Method, c.Request.URL, n))
		default:
			slog.Debug(fmt.Sprintf(formatLogger, latency, c.Request.RemoteAddr, c.status, c.Request.Method, c.Request.URL, n))
		}
	}
}

// https://www.cnblogs.com/cheyunhua/p/18049634
// https://zhuanlan.zhihu.com/p/653857076
