package whttp

import (
	"bytes"
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"

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
		return err
	}
	buf := bytes.NewBuffer(make([]byte, 0, 512))
	err = t.Execute(buf, data)
	if err != nil {
		return err
	}
	mf.dict.Store(key, buf)
	return nil

}

func (mf *MemoryFile) WriteTo(key string, w io.Writer) (int, error) {
	v, ok := mf.dict.Load(key)
	if !ok {
		return 0, errors.New("writeTo: key no found: " + key)
	}
	return w.Write(v.(*bytes.Buffer).Bytes())
}

// DelFile 删除缓存文件
func (mf *MemoryFile) DeleteFile(file string) {
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
	r.mux.HandleFunc(file, r.warp(group, "GET"))
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
			r.mux.HandleFunc(np[:len(np)-5], r.warp(append(group, fn), "GET"))
			return nil
		}
		e := r.CacheFile(np, mf, group...)
		return e
	})
	return err
}
