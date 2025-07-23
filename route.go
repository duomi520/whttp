package whttp

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
)

// DefaultMarshal 缺省JSON编码器
var DefaultMarshal func(any) ([]byte, error) = json.Marshal

// DefaultUnmarshal 缺省JSON解码器
var DefaultUnmarshal func([]byte, any) error = json.Unmarshal

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
	// 实例级中间件
	middlewares []func(*HTTPContext)
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
			if r.debugMode {
				c.Error(fmt.Sprintf("%s[%d]:%s", runtime.FuncForPC(pc).Name(), l, err.Error()))
			}
			r.logger.Error("response write error", "error", err, "funcForPC", runtime.FuncForPC(pc).Name(), "line", l, "written_bytes", n)
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
	r.middlewares = append(r.middlewares, g...)
}

// GET 注册GET方法
func (r *WRoute) GET(pattern string, fn ...func(*HTTPContext)) {
	r.Mux.HandleFunc(pattern, r.wrap(fn, "GET"))
}

// POST 注册POST方法
func (r *WRoute) POST(pattern string, fn ...func(*HTTPContext)) {
	r.Mux.HandleFunc(pattern, r.wrap(fn, "POST"))
}

// PUT 注册PUT方法
func (r *WRoute) PUT(pattern string, fn ...func(*HTTPContext)) {
	r.Mux.HandleFunc(pattern, r.wrap(fn, "PUT"))
}

// DELETE 注册DELETE方法
func (r *WRoute) DELETE(pattern string, fn ...func(*HTTPContext)) {
	r.Mux.HandleFunc(pattern, r.wrap(fn, "DELETE"))
}

// HEAD 注册HEAD方法
func (r *WRoute) HEAD(pattern string, fn ...func(*HTTPContext)) {
	r.Mux.HandleFunc(pattern, r.wrap(fn, "HEAD"))
}

// wrap 封装
func (r *WRoute) wrap(g []func(*HTTPContext), method string) func(http.ResponseWriter, *http.Request) {
	for _, v := range g {
		if v == nil {
			panic("中间件不为nil")
		}
	}
	if r.middlewares != nil {
		g = append(r.middlewares, g...)
	}
	return func(rw http.ResponseWriter, req *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				const stackSize = 4096
				buf := make([]byte, stackSize)
				lenght := runtime.Stack(buf, false)
				r.logger.Error("panic recovered", "error", v, "stack", string(buf[:lenght]))
				if r.debugMode {
					rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
					rw.WriteHeader(http.StatusInternalServerError)
					rw.Write(buf[:lenght])
				} else {
					http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
				}
			}
		}()
		if !strings.EqualFold(method, req.Method) {
			rw.Header().Set("Allow", method)
			rw.WriteHeader(http.StatusMethodNotAllowed)
			io.WriteString(rw, "method not allowed")
			r.logger.Error(fmt.Sprintf("not a %s request", req.Method), "url", req.URL)
			return
		}
		c := HTTPContextPool.Get().(*HTTPContext)
		c.index = 0
		c.chain = g
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
		c.reset()
		HTTPContextPool.Put(c)
	}
}

// Static 将指定目录下的静态文件映射到URL路径中
func (r *WRoute) Static(relativePath, file string, group ...func(*HTTPContext)) {
	// 验证路径安全性
	if strings.Contains(relativePath, "..") || strings.Contains(file, "..") {
		panic("路径包含非法字符 '..'")
	}
	for _, v := range group {
		if v == nil {
			panic("中间件不为nil")
		}
	}
	file = filepath.Clean(file)
	fn := func(c *HTTPContext) {
		c.File(file)
	}
	r.GET(relativePath, append(group, fn)...)
}

// StaticFS 静态文件目录服务
func (r *WRoute) StaticFS(dir string, group ...func(*HTTPContext)) {
	for _, v := range group {
		if v == nil {
			panic("中间件不为nil")
		}
	}
	dir = "/" + filepath.ToSlash(dir) + "/"
	fn := func(c *HTTPContext) {
		http.FileServer(http.Dir(".")).ServeHTTP(c.Writer, c.Request)
	}
	r.GET(dir, append(group, fn)...)
}

// https://mp.weixin.qq.com/s/n-kU6nwhOH6ouhufrP_1kQ
// https://zhuanlan.zhihu.com/p/679527662
