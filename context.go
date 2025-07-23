package whttp

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/duomi520/utils"
)

// H map[string]any 缩写
type H map[string]any

// HTTPContext 上下文
type HTTPContext struct {
	index   int
	chain   []func(*HTTPContext)
	mu      sync.RWMutex
	keys    utils.MetaDict[any]
	Writer  http.ResponseWriter
	Request *http.Request
	Flush   func() (int, error)
	route   *WRoute
}

func (c *HTTPContext) reset() {
	c.index = -1
	c.chain = nil
	if c.keys.Len() > 0 {
		c.keys.Key = c.keys.Key[:0]
		c.keys.Value = c.keys.Value[:0]
	}
	c.Writer = nil
	c.Request = nil
	c.Flush = nil
	c.route = nil
}

var HTTPContextPool = sync.Pool{
	New: func() any {
		return &HTTPContext{}
	},
}

func (c *HTTPContext) Debug(msg string, args ...any) {
	if c.route != nil && c.route.logger != nil {
		c.route.logger.Debug(msg, args...)
	}
}
func (c *HTTPContext) Info(msg string, args ...any) {
	if c.route != nil && c.route.logger != nil {
		c.route.logger.Info(msg, args...)
	}
}
func (c *HTTPContext) Warn(msg string, args ...any) {
	if c.route != nil && c.route.logger != nil {
		c.route.logger.Warn(msg, args...)
	}
}
func (c *HTTPContext) Error(msg string, args ...any) {
	if c.route != nil && c.route.logger != nil {
		c.route.logger.Error(msg, args...)
	}
}

// Set 上下文key-value值
func (c *HTTPContext) Set(k string, v any) {
	c.mu.Lock()
	c.keys = c.keys.Set(k, v)
	c.mu.Unlock()
}

// Get 上下文key-value值
func (c *HTTPContext) Get(k string) (v any, b bool) {
	c.mu.RLock()
	v, b = c.keys.Get(k)
	c.mu.RUnlock()
	return
}

// Del 上下文key-value值
func (c *HTTPContext) Del(k string) {
	c.mu.Lock()
	c.keys = c.keys.Del(k)
	c.mu.Unlock()
}

// BindJSON 绑定JSON数据
func (c *HTTPContext) BindJSON(v any) error {
	buf, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	if err != nil {
		return fmt.Errorf("reread body failed: %w", err)
	}
	err = DefaultUnmarshal(buf, v)
	if err != nil {
		return fmt.Errorf("unmarshal %v failed: %w", v, err)
	}
	return nil
}

// String 带有状态码的纯文本响应
func (c *HTTPContext) String(status int, msg string) {
	c.Flush = func() (int, error) {
		c.Writer.WriteHeader(status)
		return io.WriteString(c.Writer, msg)
	}
}

// JSON 带有状态码的JSON 数据
func (c *HTTPContext) JSON(status int, v any) {
	d, err := DefaultMarshal(v)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Flush = func() (int, error) {
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.WriteHeader(status)
		return c.Writer.Write(d)
	}
}

// Blob
func (c *HTTPContext) Blob(status int, contentType string, data []byte) {
	c.Flush = func() (int, error) {
		c.Writer.Header().Set("Content-Type", contentType)
		c.Writer.WriteHeader(status)
		return c.Writer.Write(data)
	}
}

// File 将静态文件返回给客户端
func (c *HTTPContext) File(filepath string) {
	c.Flush = func() (int, error) {
		var fi os.FileInfo
		var err error
		if fi, err = os.Stat(filepath); os.IsNotExist(err) {
			return 0, fmt.Errorf("file not found: %s", filepath)
		}
		//ETag 缓存验证
		etag := fmt.Sprintf("%x", fi.ModTime().UnixNano())
		c.Writer.Header().Set("ETag", etag)
		http.ServeFile(c.Writer, c.Request, filepath)
		return 0, nil
	}
}

// Render 渲染模板
func (c *HTTPContext) Render(status int, name string, v any) {
	if c.route.renderer != nil {
		c.Flush = func() (int, error) {
			c.Writer.WriteHeader(status)
			c.Writer.Header().Set("Content-Type", "text/html")
			return 0, c.route.renderer.ExecuteTemplate(c.Writer, name, v)
		}
	} else {
		c.Error("renderer not initialized", nil)
		c.String(http.StatusInternalServerError, "server error")
	}
}

// Next 下一个
func (c *HTTPContext) Next() {
	c.index++
	if c.index >= 0 && c.index < len(c.chain) {
		c.chain[c.index](c)
	}
}

// https://www.cnblogs.com/f-ck-need-u/p/10035801.html
