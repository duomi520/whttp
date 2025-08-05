package whttp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/duomi520/utils"
)

// H map[string]any 缩写
type H map[string]any

// HTTPContext 上下文
type HTTPContext struct {
	index                int
	status               int
	chain                []func(*HTTPContext)
	mu                   sync.RWMutex
	keys                 utils.MetaDict[any]
	Writer               http.ResponseWriter
	Request              *http.Request
	HookBeforWriteHeader []func(*bytes.Buffer) *bytes.Buffer
	route                *WRoute
}

func (c *HTTPContext) reset() {
	c.index = 0
	c.status = 0
	if len(c.chain) > 0 {
		c.chain = c.chain[:0]
	}
	if c.keys.Len() > 0 {
		c.keys.Key = c.keys.Key[:0]
		c.keys.Value = c.keys.Value[:0]
	}
	c.Writer = nil
	c.Request = nil
	if len(c.HookBeforWriteHeader) > 0 {
		c.HookBeforWriteHeader = c.HookBeforWriteHeader[:0]
	}
	c.route = nil
}

var HTTPContextPool = sync.Pool{
	New: func() any {
		return &HTTPContext{
			index:  0,
			status: 0,
		}
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

func (c *HTTPContext) write(status int, b []byte) (n int, err error) {
	c.status = status
	if len(c.HookBeforWriteHeader) > 0 {
		buf := c.route.pool.AllocBuffer()
		defer c.route.pool.FreeBuffer(buf)
		n, err = buf.Write(b)
		if err != nil {
			return
		}
		for i := len(c.HookBeforWriteHeader) - 1; i > -1; i-- {
			buf = c.HookBeforWriteHeader[i](buf)
			if buf == nil {
				return 0, errors.New("hookBeforWriteHeader return nil")
			}
		}
		c.Writer.WriteHeader(c.status)
		var size int64
		size, err = io.Copy(c.Writer, buf)
		if err != nil {
			c.Warn("write", "error", err.Error())
			err = nil
		}
		n = int(size)
	} else {
		c.Writer.WriteHeader(c.status)
		n, err = c.Writer.Write(b)
	}
	c.route.HookIOWriteError(c, n, err)
	return
}

// String 带有状态码的纯文本响应
func (c *HTTPContext) String(status int, msg string) {
	c.write(status, []byte(msg))
}

// JSON 带有状态码的JSON 数据
func (c *HTTPContext) JSON(status int, v any) {
	data, err := DefaultMarshal(v)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.write(status, data)
}

// Blob
func (c *HTTPContext) Blob(status int, contentType string, data []byte) {
	c.Writer.Header().Set("Content-Type", contentType)
	c.write(status, data)
}

// File 将静态文件返回给客户端
func (c *HTTPContext) File(path string) {
	// 清理路径防止目录遍历
	path = filepath.Clean(path)
	// 打开请求的文件
	f, err := os.Open(path)
	if err != nil {
		c.Error("file", "error", err.Error())
		c.write(http.StatusNotFound, []byte("404 page not found"))
		return
	}
	defer f.Close()
	// 获取文件信息
	stat, err := f.Stat()
	if err != nil {
		c.Error("file", "error", err.Error())
		c.write(http.StatusForbidden, []byte("forbidden"))
		return
	}
	// 处理目录请求
	if stat.IsDir() {
		c.Error("file", "error", "directory is not supported")
		c.write(http.StatusForbidden, []byte("forbidden"))
		return
	}
	// 服务文件内容
	// 设置 Last-Modified 头
	c.Writer.Header().Set("Last-Modified", stat.ModTime().UTC().Format(http.TimeFormat))
	// 处理 If-Modified-Since 条件请求
	IfModifiedSince := c.Request.Header.Get("If-Modified-Since")
	if len(IfModifiedSince) > 0 {
		if t, err := time.Parse(http.TimeFormat, IfModifiedSince); err == nil {
			if stat.ModTime().UTC().Unix() <= t.Unix() {
				c.write(http.StatusNotModified, []byte(""))
				return
			}
		} else {
			c.Warn("file", "error", err.Error())
		}
	}
	// 设置内容类型
	ctype := mime.TypeByExtension(filepath.Ext(stat.Name()))
	if len(ctype) > 0 {
		c.Writer.Header().Set("Content-Type", ctype)
	}
	c.status = http.StatusOK
	if len(c.HookBeforWriteHeader) > 0 {
		buf := c.route.pool.AllocBuffer()
		defer c.route.pool.FreeBuffer(buf)
		n64, err := io.Copy(buf, f)
		if err != nil {
			c.route.HookIOWriteError(c, int(n64), err)
			return
		}
		for i := len(c.HookBeforWriteHeader) - 1; i > -1; i-- {
			buf = c.HookBeforWriteHeader[i](buf)
			if buf == nil {
				c.route.HookIOWriteError(c, 0, errors.New("file hookBeforWriteHeader return nil"))
				return
			}
		}
		c.Writer.WriteHeader(c.status)
		n64, err = io.Copy(c.Writer, buf)
		c.route.HookIOWriteError(c, int(n64), err)
	} else {
		c.Writer.WriteHeader(c.status)
		n64, err := io.Copy(c.Writer, f)
		c.route.HookIOWriteError(c, int(n64), err)
	}
}

// Render 渲染模板
func (c *HTTPContext) Render(status int, name string, v any) {
	if c.route.renderer != nil {
		c.Writer.Header().Set("Content-Type", "text/html")
		c.status = status
		if len(c.HookBeforWriteHeader) > 0 {
			buf := c.route.pool.AllocBuffer()
			defer c.route.pool.FreeBuffer(buf)
			err := c.route.renderer.ExecuteTemplate(buf, name, v)
			if err != nil {
				c.Error("render", "error", err.Error())
				c.write(http.StatusInternalServerError, []byte("server error"))
				return
			}
			for i := len(c.HookBeforWriteHeader) - 1; i > -1; i-- {
				buf = c.HookBeforWriteHeader[i](buf)
				if buf == nil {
					c.route.HookIOWriteError(c, 0, errors.New("render hookBeforWriteHeader return nil"))
					return
				}
			}
			var n64 int64
			c.Writer.WriteHeader(c.status)
			n64, err = io.Copy(c.Writer, buf)
			c.route.HookIOWriteError(c, int(n64), err)
		} else {
			c.Writer.WriteHeader(c.status)
			err := c.route.renderer.ExecuteTemplate(c.Writer, name, v)
			if err != nil {
				c.Error("render", "error", err.Error())
				c.write(http.StatusInternalServerError, []byte("server error"))
			}
		}
	} else {
		c.Error("renderer not initialized", nil)
		c.write(http.StatusInternalServerError, []byte("server error"))
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
