package whttp

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"compress/gzip"
	"html/template"
	"io"
	"sync"

	"github.com/duomi520/utils"
)

type MemoryFile struct {
	dict sync.Map
}

// CacheTemplate 缓存模板
func (mf *MemoryFile) CacheTemplate(parseFiles, key string, data any) error {
	t, err := template.ParseFiles(parseFiles)
	if err != nil {
		return fmt.Errorf("缓存模板(step1)失败: %w", err)
	}
	buf := bytes.NewBuffer(make([]byte, 0, 512))
	err = t.Execute(buf, data)
	if err != nil {
		return fmt.Errorf("缓存模板(step2)失败: %w", err)
	}
	mf.dict.Store(key, buf)
	return nil
}

// ConfirmGZIP
func ConfirmGZIP(req *http.Request) bool {
	if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") &&
		strings.Contains(req.Header.Get("Connection"), "Upgrade") &&
		strings.Contains(req.Header.Get("Accept"), "text/event-stream") {
		return true
	}
	return false
}

// CacheTemplateGZIP 缓存GZIP压缩模板
func (mf *MemoryFile) CacheTemplateGZIP(level int, parseFiles, key string, data any) error {
	t, err := template.ParseFiles(parseFiles)
	if err != nil {
		return fmt.Errorf("缓存GZIP压缩模板(step1)失败: %w", err)
	}
	//TODO 减少内存浪费
	buf := bytes.NewBuffer(make([]byte, 0, 512))
	gz, err := gzip.NewWriterLevel(buf, level)
	if err != nil {
		return fmt.Errorf("缓存GZIP压缩模板(step2)失败: %w", err)
	}
	defer gz.Close()
	err = t.Execute(gz, data)
	if err != nil {
		return fmt.Errorf("缓存GZIP压缩模板(step3)失败: %w", err)
	}
	mf.dict.Store(key, buf)
	return nil
}

func (mf *MemoryFile) WriteTo(key string, w io.Writer) (int, error) {
	v, ok := mf.dict.Load(key)
	if !ok {
		return 0, errors.New("key no found: " + key)
	}
	return w.Write(v.(*bytes.Buffer).Bytes())
}

// DeleteCacheFile 删除缓存文件
func (mf *MemoryFile) DeleteCacheFile(file string) {
	mf.dict.Delete(file)
}

func (mf *MemoryFile) cacheFile(file string) error {
	data, err := os.ReadFile("." + file)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(data)
	mf.dict.Store(file, buf)
	return nil
}

// CacheFile 缓存单个静态文件
func (r *WRoute) CacheFile(file string, mf *MemoryFile, group ...func(*HTTPContext)) error {
	err := mf.cacheFile(file)
	if err != nil {
		return err
	}
	fn := func(c *HTTPContext) {
		_, err := mf.WriteTo(file, c.Writer)
		if err != nil {
			c.String(utils.StatusInternalServerError, err.Error())
		}
	}
	group = append(group, fn)
	r.Mux.HandleFunc(file, r.warp(group, "GET"))
	return nil
}

// CacheFS 当客户端以GET方法请求dir目录时，将返回缓存中的文件

func (r *WRoute) CacheFS(dir string, mf *MemoryFile, group ...func(*HTTPContext)) error {
	//遍历目录，读出文件名
	err := filepath.Walk(dir, func(p string, fi os.FileInfo, err error) error {
		if fi == nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		np := "/" + strings.ReplaceAll(p, "\\", "/")
		if strings.EqualFold(path.Ext(fi.Name()), ".tmpl") {
			//模板的匹配去掉后缀
			fn := func(c *HTTPContext) {
				_, err := mf.WriteTo(np[:len(np)-5], c.Writer)
				if err != nil {
					c.String(utils.StatusInternalServerError, err.Error())
				}
			}
			r.Mux.HandleFunc(np[:len(np)-5], r.warp(append(group, fn), "GET"))
			return nil
		}
		e := r.CacheFile(np, mf, group...)
		return e
	})
	return err
}
