package whttp

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime"
	"strings"
)

// DefaultMarshal 缺省JSON编码器
var DefaultMarshal func(any) ([]byte, error) = json.Marshal

// DefaultUnmarshal 缺省JSON解码器
var DefaultUnmarshal func([]byte, any) error = json.Unmarshal

// globalMiddleware 全局中间件
var globalMiddleware []func(*HTTPContext)

// Renderer 模板渲染接口
type Renderer interface {
	ExecuteTemplate(io.Writer, string, any) error
}

// WRoute 路由
type WRoute struct {
	debugMode bool
	Mux       *http.ServeMux
	//模版
	renderer Renderer
	//HTTPContext String、JSON、Render IO Write时错误的处理函数
	HookIOWriteError func(*HTTPContext, int, error)
	//logger
	logger *slog.Logger
}

// NewRoute 新建
func NewRoute(l *slog.Logger) *WRoute {
	r := WRoute{}
	r.Mux = http.NewServeMux()
	r.HookIOWriteError = func(c *HTTPContext, n int, err error) {
		if err != nil {
			pc, _, l, _ := runtime.Caller(2)
			c.Error(fmt.Sprintf("%s[%d]:%s", runtime.FuncForPC(pc).Name(), l, err.Error()))
		}
	}
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

func (r *WRoute) SetRenderer(s Renderer) {
	r.renderer = s
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
		if c.keys.Len() > 0 {
			c.keys.Key = c.keys.Key[:0]
			c.keys.Value = c.keys.Value[:0]
		}
		c.Writer = rw
		c.Request = req
		c.Flush = nil
		c.route = r
		c.chain[0](c)
		//数据写入下层
		if c.Flush != nil {
			n, err := c.Flush()
			r.HookIOWriteError(c, n, err)
		}
		HTTPContextPool.Put(c)
	}
}

// https://mp.weixin.qq.com/s/n-kU6nwhOH6ouhufrP_1kQ
// https://zhuanlan.zhihu.com/p/679527662
