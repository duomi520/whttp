package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/duomi520/utils"
	"github.com/gorilla/mux"
)

type HTTPGroup = []func(*HTTPContext)

//HTTPContext 上下文
type HTTPContext struct {
	index   int
	chain   HTTPGroup
	Vars    map[string]string
	Writer  http.ResponseWriter
	Request *http.Request
	route   *WRoute
}

//Params 请求参数
func (c *HTTPContext) Params(s string) string {
	if v, ok := c.Vars[s]; ok {
		return v
	}
	return c.Request.FormValue(s)
}

//BindJSON 绑定JSON数据
func (c *HTTPContext) BindJSON(v any) error {
	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return fmt.Errorf("readAll faile: %w", err)
	}
	err = json.Unmarshal(buf, v)
	if err != nil {
		return fmt.Errorf("unmarshal %v faile: %w", v, err)
	}
	err = c.route.validatorStruct(v)
	if err != nil {
		return fmt.Errorf("validator %v faile: %w", v, err)
	}
	return nil
}

//String 带有状态码的纯文本响应
func (c *HTTPContext) String(status int, msg string) {
	c.Writer.WriteHeader(status)
	io.WriteString(c.Writer, msg)
}

//JSON 带有状态码的JSON 数据
func (c *HTTPContext) JSON(status int, v any) {
	d, err := json.Marshal(v)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)
	c.Writer.Write(d)
}

//Next 下一个
func (c *HTTPContext) Next() {
	c.index++
	c.chain[c.index](c)
}

//WRoute w
type WRoute struct {
	//mux
	router *mux.Router
	//validator
	validatorVar    func(any, string) error
	validatorStruct func(any) error
	//logger
	logger    utils.ILogger
	DebugMode bool
}

//NewRoute 新建
func NewRoute(v utils.IValidator, log utils.ILogger) *WRoute {
	r := WRoute{}
	r.router = mux.NewRouter()
	if v == nil {
		panic("Validator is nil")
	}
	r.validatorVar = v.Var
	r.validatorStruct = v.Struct
	if log == nil {
		panic("Logger is nil")
	}
	r.logger = log
	return &r
}

//PathPrefix 前缀
func (r *WRoute) PathPrefix(tpl string) {
	r.router.PathPrefix(tpl).Handler(http.DefaultServeMux)
}

//HandleFunc 处理
func (r *WRoute) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	r.router.HandleFunc(pattern, handler).Methods("GET")
}

//GET g
func (r *WRoute) GET(g HTTPGroup, url string, fn func(*HTTPContext)) {
	r.router.HandleFunc(url, r.Warp(g, fn)).Methods("GET")
}

//POST p
func (r *WRoute) POST(g HTTPGroup, url string, fn func(*HTTPContext)) {
	r.router.HandleFunc(url, r.Warp(g, fn)).Methods("POST")
}

//DELETE d
func (r *WRoute) DELETE(g HTTPGroup, url string, fn func(*HTTPContext)) {
	r.router.HandleFunc(url, r.Warp(g, fn)).Methods("DELETE")
}

//Warp 封装
func (r *WRoute) Warp(g HTTPGroup, fn func(*HTTPContext)) func(http.ResponseWriter, *http.Request) {
	chain := make(HTTPGroup, len(g)+1)
	copy(chain, g)
	chain[len(g)] = fn
	return func(rw http.ResponseWriter, req *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				buf := make([]byte, 4096)
				lenght := runtime.Stack(buf, false)
				r.logger.Error(fmt.Sprintf("WRoute.warp %v \n%s", v, buf[:lenght]))
				if r.DebugMode {
					rw.Write([]byte("\n"))
					rw.Write(buf[:lenght])
				}
			}
		}()
		c := &HTTPContext{chain: chain, Writer: rw, Request: req, route: r}
		c.Vars = mux.Vars(req)
		c.chain[0](c)
	}
}

//HTTPMiddleware 中间件
func HTTPMiddleware(m ...func(*HTTPContext)) HTTPGroup {
	return m
}

//ValidatorMiddleware 输入参数验证
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
			if err := c.route.validatorVar(c.Params(s0[i]), s1[i]); err != nil {
				c.String(http.StatusBadRequest, fmt.Sprintf("validate %s %s faile: %s", s0[i], s1[i], err.Error()))
				return
			}
		}
		c.Next()
	}
}

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

// https://github.com/julienschmidt/httprouter
// https://mp.weixin.qq.com/s/9P1AV6d_Cc4pH9DNJeEHHg
