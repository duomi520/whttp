package whttp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sync"

	"github.com/duomi520/utils"
	"github.com/julienschmidt/httprouter"
)

type Vars interface {
	ByName(string) string
}

// HTTPContext 上下文
type HTTPContext struct {
	index   int
	chain   []func(*HTTPContext)
	mu      sync.RWMutex
	keys    map[string]any
	vars    Vars
	Writer  http.ResponseWriter
	Request *http.Request
	route   *WRoute
}

var HTTPContextPool = sync.Pool{
	New: func() interface{} {
		return &HTTPContext{}
	},
}

// Set 上下文key-value值
func (c *HTTPContext) Set(k string, v any) {
	c.mu.Lock()
	if c.keys == nil {
		c.keys = make(map[string]any)
	}
	c.keys[k] = v
	c.mu.Unlock()
}

// Get 上下文key-value值
func (c *HTTPContext) Get(k string) (v any, b bool) {
	if c.keys == nil {
		return
	}
	c.mu.RLock()
	v, b = c.keys[k]
	c.mu.RUnlock()
	return
}

// Params 请求参数
func (c *HTTPContext) Params(s string) string {
	return c.vars.ByName(s)
}

// BindJSON 绑定JSON数据
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

// String 带有状态码的纯文本响应
func (c *HTTPContext) String(status int, msg string) {
	c.Writer.WriteHeader(status)
	io.WriteString(c.Writer, msg)
}

// JSON 带有状态码的JSON 数据
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

// Next 下一个
func (c *HTTPContext) Next() {
	c.index++
	c.chain[c.index](c)
}

// WRoute 路由
type WRoute struct {
	DebugMode bool
	router    *httprouter.Router
	//validator
	validatorVar    func(any, string) error
	validatorStruct func(any) error
	//logger
	logger utils.ILogger
}

// NewRoute 新建
func NewRoute(v utils.IValidator, log utils.ILogger) *WRoute {
	r := WRoute{}
	r.router = httprouter.New()
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

// GET g
func (r *WRoute) GET(pattern string, fn ...func(*HTTPContext)) {
	r.router.GET(pattern, r.warp(fn))
}

// POST p
func (r *WRoute) POST(pattern string, fn ...func(*HTTPContext)) {
	r.router.POST(pattern, r.warp(fn))
}

// DELETE d
func (r *WRoute) DELETE(pattern string, fn ...func(*HTTPContext)) {
	r.router.DELETE(pattern, r.warp(fn))
}

// warp 封装
func (r *WRoute) warp(g []func(*HTTPContext)) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
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
		c.chain = g
		c.vars = p
		c.Writer = rw
		c.Request = req
		c.route = r
		c.chain[0](c)
		HTTPContextPool.Put(c)
	}
}

// https://mp.weixin.qq.com/s/n-kU6nwhOH6ouhufrP_1kQ
// https://www.cnblogs.com/cheyunhua/p/15545261.html
