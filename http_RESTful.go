package whttp

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime"
	"strings"
	"sync"

	"github.com/duomi520/utils"
)

// DefaultMarshal 缺省JSON编码器
var DefaultMarshal func(any) ([]byte, error) = json.Marshal

// DefaultUnmarshal 缺省JSON解码器
var DefaultUnmarshal func([]byte, any) error = json.Unmarshal

// globalMiddleware 全局中间件
var globalMiddleware []func(*HTTPContext)

// HTTPContext 上下文
type HTTPContext struct {
	index   int
	chain   []func(*HTTPContext)
	mu      sync.RWMutex
	keys    map[string]any
	Writer  http.ResponseWriter
	Request *http.Request
	route   *WRoute
}

var HTTPContextPool = sync.Pool{
	New: func() interface{} {
		return &HTTPContext{}
	},
}

func (c *HTTPContext) Debug(msg string, args ...any) {
	c.route.logger.Debug(msg, args...)
}
func (c *HTTPContext) Info(msg string, args ...any) {
	c.route.logger.Info(msg, args...)
}
func (c *HTTPContext) Warn(msg string, args ...any) {
	c.route.logger.Warn(msg, args...)
}
func (c *HTTPContext) Error(msg string, args ...any) {
	c.route.logger.Error(msg, args...)
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

// BindJSON 绑定JSON数据
func (c *HTTPContext) BindJSON(v any) error {
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return fmt.Errorf("readAll faile: %w", err)
	}
	err = DefaultUnmarshal(buf, v)
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
	d, err := DefaultMarshal(v)
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
	if c.index < len(c.chain) {
		c.chain[c.index](c)
	}
}

// File 将静态文件直接返回给客户端
func (c *HTTPContext) File(filepath string) {
	http.ServeFile(c.Writer, c.Request, filepath)
}

// WRoute 路由
type WRoute struct {
	debugMode bool
	Mux       *http.ServeMux
	//validator
	validatorVar    func(any, string) error
	validatorStruct func(any) error
	//logger
	logger *slog.Logger
}

// NewRoute 新建
func NewRoute(v utils.IValidator, l *slog.Logger) *WRoute {
	r := WRoute{}
	r.Mux = http.NewServeMux()
	if v == nil {
		panic("Validator is nil")
	}
	r.validatorVar = v.Var
	r.validatorStruct = v.Struct
	if l == nil {
		r.logger = slog.Default()
	} else {
		r.logger = l
	}
	return &r
}

func (r *WRoute) SetDebugMode(b bool) {
	r.debugMode = b
}

// Use 全局中间件 需放在最前面
func (r *WRoute) Use(g ...func(*HTTPContext)) {
	for _, v := range g {
		if v == nil {
			panic("中间件不为nil")
		}
	}
	globalMiddleware = append(globalMiddleware, g...)
}

// GET g
func (r *WRoute) GET(pattern string, fn ...func(*HTTPContext)) {
	r.Mux.HandleFunc(pattern, r.warp(fn, "GET"))
}

// POST p
func (r *WRoute) POST(pattern string, fn ...func(*HTTPContext)) {
	r.Mux.HandleFunc(pattern, r.warp(fn, "POST"))
}

// PUT p
func (r *WRoute) PUT(pattern string, fn ...func(*HTTPContext)) {
	r.Mux.HandleFunc(pattern, r.warp(fn, "PUT"))
}

// DELETE d
func (r *WRoute) DELETE(pattern string, fn ...func(*HTTPContext)) {
	r.Mux.HandleFunc(pattern, r.warp(fn, "DELETE"))
}

// warp 封装
func (r *WRoute) warp(g []func(*HTTPContext), method string) func(http.ResponseWriter, *http.Request) {
	for _, v := range g {
		if v == nil {
			panic("中间件不为nil")
		}
	}
	if globalMiddleware != nil {
		g = append(globalMiddleware, g...)
	}
	return func(rw http.ResponseWriter, req *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				buf := make([]byte, 4096)
				lenght := runtime.Stack(buf, false)
				r.logger.Error(fmt.Sprintf(" %v \n%s", v, buf[:lenght]))
				if r.debugMode {
					rw.WriteHeader(http.StatusInternalServerError)
					rw.Write(buf[:lenght])
				}
			}
		}()
		if !strings.EqualFold(method, req.Method) {
			r.logger.Error(fmt.Sprintf("not a %s request", req.Method))
			rw.WriteHeader(http.StatusMethodNotAllowed)
			io.WriteString(rw, "method not allowed")
			return
		}
		c := HTTPContextPool.Get().(*HTTPContext)
		c.index = 0
		c.chain = g
		c.Writer = rw
		c.Request = req
		c.route = r
		c.chain[0](c)
		HTTPContextPool.Put(c)
	}
}

// https://mp.weixin.qq.com/s/n-kU6nwhOH6ouhufrP_1kQ
// https://zhuanlan.zhihu.com/p/679527662
