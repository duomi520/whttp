package whttp
import (
	"fmt"
	"net/http"
	"time"
)

var formatLogger string = "| %13d | %15s | %5d | %7s | %s | %s "

//LoggerMiddleware 日志
func LoggerMiddleware() func(*HTTPContext) {
	var startTime time.Time
	var rw httpLoggerResponseWriter
	return func(c *HTTPContext) {
		startTime = time.Now()
		rw.w = c.Writer
		c.Writer = &rw
		c.Next()
		if rw.status > 299 {
			if rw.err != nil {
				c.route.logger.Error(fmt.Sprintf(formatLogger, time.Since(startTime), c.Request.RemoteAddr, rw.status, c.Request.Method, c.Request.URL, rw.err.Error()))
			} else {
				c.route.logger.Warn(fmt.Sprintf(formatLogger, time.Since(startTime), c.Request.RemoteAddr, rw.status, c.Request.Method, c.Request.URL, rw.result))
			}
		} else {
			c.route.logger.Debug(fmt.Sprintf(formatLogger, time.Since(startTime), c.Request.RemoteAddr, rw.status, c.Request.Method, c.Request.URL, rw.result))
		}
	}
}

//httpLoggerResponseWriter d
type httpLoggerResponseWriter struct {
	status int
	result []byte
	err    error
	w      http.ResponseWriter
}

//Header 返回一个Header类型值
func (h *httpLoggerResponseWriter) Header() http.Header {
	return h.w.Header()
}

//WriteHeader 该方法发送HTTP回复的头域和状态码
func (h *httpLoggerResponseWriter) WriteHeader(s int) {
	h.status = s
	h.w.WriteHeader(s)
}

//Write 向连接中写入作为HTTP的一部分回复的数据
func (h *httpLoggerResponseWriter) Write(d []byte) (int, error) {
	n, err := h.w.Write(d)
	h.result = d
	h.err = err
	return n, err
}
