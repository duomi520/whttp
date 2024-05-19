package whttp

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"compress/gzip"
	"html/template"
	"io"
	"sync"

	"github.com/duomi520/utils"
)

var GZIPExcludedExtentions []string = []string{".png", ".gif", ".jpeg", ".jpg"}

type MemoryFile struct {
	dict sync.Map
}

// CacheTemplate 缓存模板
func (mf *MemoryFile) CacheTemplate(parseFiles, key string, data any) error {
	t, err := template.ParseFiles(parseFiles)
	if err != nil {
		return fmt.Errorf("缓存模板(step1)失败: %w", err)
	}
	buf := bytes.NewBuffer(make([]byte, 0, 64))
	err = t.Execute(buf, data)
	if err != nil {
		return fmt.Errorf("缓存模板(step2)失败: %w", err)
	}
	mf.dict.Store(key, buf)
	return nil
}

// CacheTemplateGZIP 缓存GZIP压缩模板
func (mf *MemoryFile) CacheTemplateGZIP(level int, parseFiles, key string, data any) error {
	t, err := template.ParseFiles(parseFiles)
	if err != nil {
		return fmt.Errorf("缓存GZIP压缩模板(step1)失败: %w", err)
	}
	buf := bytes.NewBuffer(make([]byte, 0, 64))
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

func (mf *MemoryFile) WriteGZIP(key string, w http.ResponseWriter, req *http.Request) (int, error) {
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
		return 0, errors.New("非gzip请求")
	}
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Vary", "Accept-Encoding")
	return mf.WriteTo(key, w)
}

// DeleteCacheFile 删除缓存文件
func (mf *MemoryFile) DeleteCacheFile(file string) {
	mf.dict.Delete(file)
}

// CacheFile 缓存单个静态文件
func (r *WRoute) CacheFile(file string, mf *MemoryFile, group ...func(*HTTPContext)) error {
	data, err := os.ReadFile("." + file)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(data)
	mf.dict.Store(file, buf)
	fn := func(c *HTTPContext) {
		_, err := mf.WriteTo(file, c.Writer)
		if err != nil {
			c.String(utils.StatusNotFound, err.Error())
		}
	}
	group = append(group, fn)
	r.Mux.HandleFunc(file, r.warp(group, "GET"))
	return nil
}

// CacheFileGZIP 用GZIP压缩缓存单个静态文件
func (r *WRoute) CacheFileGZIP(level int, file string, mf *MemoryFile, group ...func(*HTTPContext)) error {
	data, err := os.ReadFile("." + file)
	if err != nil {
		return fmt.Errorf("GZIP压缩缓存文件(step1)失败: %w", err)
	}
	buf := bytes.NewBuffer(make([]byte, 0, 256))
	gz, err := gzip.NewWriterLevel(buf, level)
	if err != nil {
		return fmt.Errorf("GZIP压缩缓存文件(step2)失败: %w", err)
	}
	defer gz.Close()
	_, err = gz.Write(data)
	if err != nil {
		return fmt.Errorf("GZIP压缩缓存文件(step3)失败: %w", err)
	}
	mf.dict.Store(file, buf)
	fn := func(c *HTTPContext) {
		c.Writer.Header().Set("Content-Encoding", "gzip")
		c.Writer.Header().Set("Vary", "Accept-Encoding")
		n, err := mf.WriteTo(file, c.Writer)
		if err != nil {
			c.String(utils.StatusNotFound, err.Error())
		}
		c.Writer.Header().Set("Content-Length", strconv.Itoa(n))
	}
	group = append(group, fn)
	r.Mux.HandleFunc(file, r.warp(group, "GET"))
	return nil
}

// CacheFS 当客户端以GET方法请求dir目录时，将返回缓存中的文件
func (r *WRoute) CacheFS(dir string, mf *MemoryFile, group ...func(*HTTPContext)) error {
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
		np := "/" + strings.ReplaceAll(p, "\\", "/")
		return r.CacheFile(np, mf, group...)
	})
}

// CacheFSGZIP 当客户端以GET方法请求dir目录时，将返回缓存中的文件,缓存中文件GZIP压缩
func (r *WRoute) CacheFSGZIP(level int, dir string, mf *MemoryFile, group ...func(*HTTPContext)) error {
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
		np := "/" + strings.ReplaceAll(p, "\\", "/")
		//后缀在GZIPExcludedExtentions中的不压缩
		for _, v := range GZIPExcludedExtentions {
			if strings.EqualFold(path.Ext(fi.Name()), v) {
				return r.CacheFile(np, mf, group...)
			}
		}
		return r.CacheFileGZIP(level, np, mf, group...)
	})
}
