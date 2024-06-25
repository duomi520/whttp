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
	var h httpLoggerResponseWriter
	return func(c *HTTPContext) {
		h.startTime = time.Now()
		h.r = c.Request
		h.w = c.Writer
		c.Writer = &h
		c.Next()
		if c.Flush == nil {
			h.log(0, nil)
		}
	}
}

// httpLoggerResponseWriter d
type httpLoggerResponseWriter struct {
	startTime time.Time
	status    int
	r         *http.Request
	w         http.ResponseWriter
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
	h.log(n, err)
	return n, err
}
func (h *httpLoggerResponseWriter) log(n int, err error) {
	latency := time.Since(h.startTime)
	if latency > time.Minute {
		latency = latency.Truncate(time.Second)
	}
	if h.status > 299 {
		if err != nil {
			slog.Error(fmt.Sprintf(formatLogger, latency, h.r.RemoteAddr, h.status, h.r.Method, h.r.URL, err.Error()))
		} else {
			slog.Warn(fmt.Sprintf(formatLogger, latency, h.r.RemoteAddr, h.status, h.r.Method, h.r.URL, strconv.Itoa(n)))
		}
	} else {
		slog.Debug(fmt.Sprintf(formatLogger, latency, h.r.RemoteAddr, h.status, h.r.Method, h.r.URL, strconv.Itoa(n)))
	}
}

// https://www.cnblogs.com/cheyunhua/p/18049634
// https://zhuanlan.zhihu.com/p/653857076
