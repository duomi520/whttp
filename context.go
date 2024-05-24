package whttp

import (
	"fmt"
	"io"
	"net/http"
	"sync"
)

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
