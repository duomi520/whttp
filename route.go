package whttp

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/duomi520/utils"
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
	//HTTPContext String、JSON、Render、File IO Write时错误的处理函数
	HookIOWriteError func(*HTTPContext, int, error)
	// 实例级中间件
	middlewares []func(*HTTPContext)
	pool        *utils.Pool
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
			r.logger.Error("hookIOWriteError", "error", err, "funcForPC", runtime.FuncForPC(pc).Name(), "line", l, "written_bytes", n)
		}
	}
	r.pool = &utils.Pool{}
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
			panic("middleware cannot be nil")
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
			panic("middleware cannot be nil")
		}
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
		c.chain = append(c.chain, r.middlewares...)
		c.chain = append(c.chain, g...)
		c.Writer = rw
		c.Request = req
		c.route = r
		c.chain[0](c)
		c.reset()
		HTTPContextPool.Put(c)
	}
}

// Static 将指定目录下的静态文件映射到URL路径中,relativePath不支持中文
func (r *WRoute) Static(relativePath, file string, group ...func(*HTTPContext)) {
	if strings.Contains(relativePath, "..") || strings.Contains(file, "..") {
		panic("path contains illegal characters '..'")
	}
	for _, v := range group {
		if v == nil {
			panic("middleware cannot be nil")
		}
	}
	fn := func(c *HTTPContext) {
		c.File(file)
	}
	r.GET(relativePath, append(group, fn)...)
}

// StaticFS 静态文件目录服务,目录名不支持中文
func (r *WRoute) StaticFS(root string, group ...func(*HTTPContext)) {
	if strings.Contains(root, "..") {
		panic("path contains illegal characters '..'")
	}
	for _, v := range group {
		if v == nil {
			panic("middleware cannot be nil")
		}
	}
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			panic("directory does not exist: " + root)
		} else {
			panic(err.Error())
		}
	}
	WalkDirFunc := func(path string, d fs.DirEntry, err error) error {
		// 返回遍历中的错误
		if err != nil {
			return err
		}
		// 跳过目录（只处理文件）
		if d.IsDir() {
			return nil
		}
		filePath, fileName := filepath.Split(path)
		// 转换路径分隔符为正斜杠（适用于URL）
		urlPath := "/" + filepath.ToSlash(filePath)
		escapeUrl := urlPath + url.QueryEscape(fileName)
		fn := func(c *HTTPContext) {
			c.File(path)
		}
		r.GET(escapeUrl, append(group, fn)...)
		return nil

	}
	// 遍历目录
	err := filepath.WalkDir(root, WalkDirFunc)
	if err != nil {
		panic(err.Error())
	}
}

// https://mp.weixin.qq.com/s/n-kU6nwhOH6ouhufrP_1kQ
// https://zhuanlan.zhihu.com/p/679527662
// https://www.cnblogs.com/schaepher/p/12831623.html
