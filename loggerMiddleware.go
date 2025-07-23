package whttp

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

var formatLogger string = "| %-13v | %-15s | %3d | %-7s | %-40s | %7d bytes "

// LoggerMiddleware 日志
func LoggerMiddleware() func(*HTTPContext) {
	var h httpLoggerResponseWriter
	return func(c *HTTPContext) {
		startTime := time.Now()
		h.r = c.Request
		h.w = c.Writer
		c.Writer = &h
		c.Next()
		flush := c.Flush
		c.Flush = func() (n int, err error) {
			if flush != nil {
				n, err = flush()
			}
			latency := time.Since(startTime)
			// 智能截断过长的延迟显示
			if latency > time.Minute {
				latency = latency.Truncate(time.Millisecond)
			}
			switch {
			case h.status >= http.StatusInternalServerError:
				slog.Error(fmt.Sprintf(formatLogger, latency, h.r.RemoteAddr, h.status, h.r.Method, h.r.URL, n))
			case h.status >= http.StatusBadRequest:
				slog.Warn(fmt.Sprintf(formatLogger, latency, h.r.RemoteAddr, h.status, h.r.Method, h.r.URL, n))
			default:
				slog.Debug(fmt.Sprintf(formatLogger, latency, h.r.RemoteAddr, h.status, h.r.Method, h.r.URL, n))
			}
			return
		}
	}
}

// httpLoggerResponseWriter d
type httpLoggerResponseWriter struct {
	status int
	r      *http.Request
	w      http.ResponseWriter
}

// Header 返回一个Header类型值
func (h *httpLoggerResponseWriter) Header() http.Header {
	return h.w.Header()
}

// WriteHeader 该方法发送HTTP回复的头域和状态码
func (h *httpLoggerResponseWriter) WriteHeader(s int) {
	h.status = s
	h.w.WriteHeader(s)
}

// Write 向连接中写入作为HTTP的一部分回复的数据
func (h *httpLoggerResponseWriter) Write(d []byte) (int, error) {
	n, err := h.w.Write(d)
	return n, err
}

// https://www.cnblogs.com/cheyunhua/p/18049634
// https://zhuanlan.zhihu.com/p/653857076
