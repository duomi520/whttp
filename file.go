package whttp

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"compress/gzip"
	"io"
	"sync"
	"time"
)

var GZIPExcludedExtentions []string = []string{".png", ".gif", ".jpeg", ".jpg"}

type FileBuffer struct {
	LastModify time.Time
	ETag       string
	Load       func() ([]byte, error)
	Buf        *bytes.Buffer
}

func (f FileBuffer) Len() int {
	if f.Buf != nil {
		return f.Buf.Len()
	}
	return 0
}

func NewFileBuffer(load func() ([]byte, error)) (FileBuffer, error) {
	data, err := load()
	if err != nil {
		return FileBuffer{}, err
	}
	fb := FileBuffer{
		LastModify: time.Now(),
		ETag:       fmt.Sprintf("%x", (md5.Sum(data))),
		Load:       load,
		Buf:        bytes.NewBuffer(data),
	}
	return fb, nil
}

type MemoryFile struct {
	sync.Map
}

func (mf *MemoryFile) GetETag(key string) string {
	v, ok := mf.Load(key)
	if !ok {
		return ""
	}
	return v.(FileBuffer).ETag
}

func (mf *MemoryFile) WriteTo(key string, w io.Writer) (int, error) {
	v, ok := mf.Load(key)
	if !ok {
		return 0, errors.New("key no found: " + key)
	}
	f := v.(FileBuffer)
	if f.Buf == nil {
		var err error
		f, err = NewFileBuffer(f.Load)
		if err != nil {
			return 0, err
		}
	}
	return w.Write(f.Buf.Bytes())
}

// StoreCacheFile 存储缓存
func (mf *MemoryFile) StoreCacheFile(key string, load func() ([]byte, error)) error {
	fb, err := NewFileBuffer(load)
	if err != nil {
		return err
	}
	mf.Store(key, fb)
	return nil
}

// FreeBuffer 释放缓存
func (mf *MemoryFile) FreeBuffer(key string) {
	v, _ := mf.Load(key)
	f := v.(FileBuffer)
	nf := FileBuffer{Load: f.Load}
	mf.Store(key, nf)
}

// DeleteCacheFile 删除缓存
func (mf *MemoryFile) DeleteCacheFile(key string) {
	mf.Delete(key)
}

func checkCacheControl(req *http.Request) bool {
	CC := req.Header.Get("Cache-Control")
	if strings.Contains(CC, "no-store") {
		return true
	}
	if strings.Contains(CC, "no-cache") {
		return true
	}
	if strings.Contains(CC, "must-revalidate") {
		return true
	}
	if strings.Contains(CC, "max-age=0") {
		return true
	}
	return false
}

// Static 将指定目录下的静态文件映射到URL路径中
func (r *WRoute) Static(relativePath, root string, group ...func(*HTTPContext)) {
	for _, v := range group {
		if v == nil {
			panic("中间件不为nil")
		}
	}
	fn := func(c *HTTPContext) {
		c.File(root)
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

// CacheFile 缓存单个静态文件
func (r *WRoute) CacheFile(file string, mf *MemoryFile, group ...func(*HTTPContext)) error {
	for _, v := range group {
		if v == nil {
			panic("中间件不为nil")
		}
	}
	load := func() ([]byte, error) {
		return os.ReadFile(file)
	}
	key := "/" + filepath.ToSlash(file)
	if err := mf.StoreCacheFile(key, load); err != nil {
		return err
	}
	fn := func(c *HTTPContext) {
		ETag := mf.GetETag(key)
		if checkCacheControl(c.Request) {
			if strings.Contains(c.Request.Header.Get("If-None-Match"), ETag) {
				c.String(http.StatusNotModified, "")
				return
			}
		}
		c.Writer.Header().Set("ETag", ETag)
		c.Writer.WriteHeader(http.StatusOK)
		_, err := mf.WriteTo(key, c.Writer)
		if err != nil {
			c.String(http.StatusNotFound, err.Error())
		}
	}
	r.GET(key, append(group, fn)...)
	return nil
}

// CacheFileGZIP 用GZIP压缩缓存单个静态文件
func (r *WRoute) CacheFileGZIP(level int, file string, mf *MemoryFile, group ...func(*HTTPContext)) error {
	for _, v := range group {
		if v == nil {
			panic("中间件不为nil")
		}
	}
	load := func() ([]byte, error) {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("GZIP压缩缓存文件(step1)失败: %w", err)
		}
		var buf bytes.Buffer
		gz, err := gzip.NewWriterLevel(&buf, level)
		if err != nil {
			return nil, fmt.Errorf("GZIP压缩缓存文件(step2)失败: %w", err)
		}
		defer gz.Close()
		_, err = gz.Write(data)
		if err != nil {
			return nil, fmt.Errorf("GZIP压缩缓存文件(step3)失败: %w", err)
		}
		err = gz.Flush()
		if err != nil {
			return nil, fmt.Errorf("GZIP压缩缓存文件(step4)失败: %w", err)
		}
		return buf.Bytes(), nil
	}
	key := "/" + filepath.ToSlash(file)
	if err := mf.StoreCacheFile(key, load); err != nil {
		return err
	}
	fn := func(c *HTTPContext) {
		ETag := mf.GetETag(key)
		if checkCacheControl(c.Request) {
			if strings.Contains(c.Request.Header.Get("If-None-Match"), ETag) {
				c.String(http.StatusNotModified, "")
				return
			}
		}
		if !strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			c.String(http.StatusNotAcceptable, "need gzip")
			return
		}
		c.Writer.Header().Set("ETag", ETag)
		c.Writer.Header().Set("Content-Encoding", "gzip")
		c.Writer.Header().Set("Vary", "Accept-Encoding")
		c.Writer.WriteHeader(http.StatusOK)
		_, err := mf.WriteTo(key, c.Writer)
		if err != nil {
			c.String(http.StatusNotFound, err.Error())
		}
	}
	r.GET(key, append(group, fn)...)
	return nil
}

// CacheFS 当客户端以GET方法请求dir目录时，将返回缓存中的文件
// 目录内文件修改，缓存并不会一起更改
// 后缀".tmpl"的文件不缓存
func (r *WRoute) CacheFS(dir string, mf *MemoryFile, group ...func(*HTTPContext)) error {
	for _, v := range group {
		if v == nil {
			panic("中间件不为nil")
		}
	}
	//遍历目录，读出文件名
	return filepath.Walk(dir, func(p string, fi os.FileInfo, err error) error {
		if fi == nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		//排除后缀".tmpl"的文件
		if strings.EqualFold(path.Ext(fi.Name()), ".tmpl") {
			return nil
		}
		return r.CacheFile(p, mf, group...)
	})
}

// CacheFSGZIP 当客户端以GET方法请求dir目录时，将返回缓存中的文件,缓存中文件GZIP压缩
// 目录内文件修改，缓存并不会一起更改
// 后缀".tmpl"的文件不缓存
// 数组中GZIPExcludedExtentions包含的后缀文件不压缩，默认设置".png", ".gif", ".jpeg", ".jpg"
func (r *WRoute) CacheFSGZIP(level int, dir string, mf *MemoryFile, group ...func(*HTTPContext)) error {
	for _, v := range group {
		if v == nil {
			panic("中间件不为nil")
		}
	}
	//遍历目录，读出文件名
	return filepath.Walk(dir, func(p string, fi os.FileInfo, err error) error {
		if fi == nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		//排除后缀".tmpl"的文件
		if strings.EqualFold(path.Ext(fi.Name()), ".tmpl") {
			return nil
		}
		//后缀在GZIPExcludedExtentions中的不压缩
		for _, v := range GZIPExcludedExtentions {
			if strings.EqualFold(path.Ext(fi.Name()), v) {
				return r.CacheFile(p, mf, group...)
			}
		}
		return r.CacheFileGZIP(level, p, mf, group...)
	})

}

// https://zhuanlan.zhihu.com/p/429777517
