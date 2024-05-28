package whttp

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

var formatLogger string = "| %13v | %15s | %5d | %7s | %s | %s "

// LoggerMiddleware 日志
func LoggerMiddleware() func(*HTTPContext) {
	var startTime time.Time
	var rw httpLoggerResponseWriter
	return func(c *HTTPContext) {
		startTime = time.Now()
		rw.w = c.Writer
		c.Writer = &rw
		c.Next()
		latency := time.Since(startTime)
		if latency > time.Minute {
			latency = latency.Truncate(time.Second)
		}
		if rw.status > 299 {
			if rw.err != nil {
				slog.Error(fmt.Sprintf(formatLogger, latency, ClientIP(c.Request), rw.status, c.Request.Method, c.Request.URL, rw.err.Error()))
			} else {
				slog.Warn(fmt.Sprintf(formatLogger, latency, ClientIP(c.Request), rw.status, c.Request.Method, c.Request.URL, strconv.Itoa(rw.length)))
			}
		} else {
			slog.Debug(fmt.Sprintf(formatLogger, latency, ClientIP(c.Request), rw.status, c.Request.Method, c.Request.URL, strconv.Itoa(rw.length)))
		}
	}
}

// httpLoggerResponseWriter d
type httpLoggerResponseWriter struct {
	status int
	length int
	err    error
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
	h.length = len(d)
	h.err = err
	return n, err
}

// https://www.cnblogs.com/cheyunhua/p/18049634
// https://zhuanlan.zhihu.com/p/653857076
