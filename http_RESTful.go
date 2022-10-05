package whttp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sync"

	"github.com/duomi520/utils"
	"github.com/gorilla/mux"
)

type HTTPGroup = []func(*HTTPContext)

//HTTPContext 上下文
type HTTPContext struct {
	index   int
	chain   HTTPGroup
	vars    map[string]string
	Writer  http.ResponseWriter
	Request *http.Request
	route   *WRoute
}

var HTTPContextPool = sync.Pool{
	New: func() interface{} {
		return &HTTPContext{}
	},
}

//Params 请求参数
func (c *HTTPContext) Params(s string) string {
	if v, ok := c.vars[s]; ok {
		return v
	}
	return c.Request.FormValue(s)
}

//BindJSON 绑定JSON数据
func (c *HTTPContext) BindJSON(v any) error {
	buf, err := io.ReadAll(c.Request.Body)
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
	DebugMode bool
	//mux
	router *mux.Router
	//validator
	validatorVar    func(any, string) error
	validatorStruct func(any) error
	//logger
	logger utils.ILogger
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
		c := HTTPContextPool.Get().(*HTTPContext)
		c.index = 0
		c.chain = chain
		c.vars = mux.Vars(req)
		c.Writer = rw
		c.Request = req
		c.route = r
		c.chain[0](c)
		HTTPContextPool.Put(c)
	}
}

//HTTPMiddleware 中间件
func HTTPMiddleware(m ...func(*HTTPContext)) HTTPGroup {
	return m
}

// https://github.com/julienschmidt/httprouter
// https://mp.weixin.qq.com/s/9P1AV6d_Cc4pH9DNJeEHHg
