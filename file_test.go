package whttp

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStatic(t *testing.T) {
	r := &WRoute{Mux: http.NewServeMux()}
	r.Static("/", "txt\\welcome.txt")
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, []byte("Welcome to the page!")) {
		t.Errorf("got %s | expected %s", string(data), "Welcome to the page!")
	}
}
func TestStaticFS(t *testing.T) {
	r := &WRoute{Mux: http.NewServeMux()}
	r.StaticFS("txt")
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	tests := [][2]string{
		{"/txt/a.txt", "a"},
		{"/txt/b.txt", "b"},
		{"/txt/1/c.txt", "c"},
		{"/txt/welcome.txt", "Welcome to the page!"},
		{"/file.tmp", "404 page not found\n"},
	}
	for i := range tests {
		resp, err := http.Get(ts.URL + tests[i][0])
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.EqualFold(tests[i][1], string(data)) {
			t.Errorf("%d expected %s got %s", i, tests[i][1], string(data))
		}
	}
}

func TestCacheFile(t *testing.T) {
	var mf MemoryFile
	welcome := "Welcome to the page!"
	r := &WRoute{Mux: http.NewServeMux()}
	err := r.CacheFile("txt\\welcome.txt", &mf)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	res, err := http.Get(ts.URL + "/txt/welcome.txt")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(res.Header.Get("ETag"), "f615cf810b8c9723d0a836dbd9df8648") {
		t.Errorf("got %s | expected f615cf810b8c9723d0a836dbd9df8648", res.Header.Get("ETag"))
	}
	data, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, []byte(welcome)) {
		t.Errorf("got %s | expected %s", string(data), welcome)
	}
}

func TestCacheFS(t *testing.T) {
	var mf MemoryFile
	r := &WRoute{Mux: http.NewServeMux()}
	err := r.CacheFS("txt", &mf)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	tests := [][2]string{
		{"/txt/a.txt", "a"},
		{"/txt/b.txt", "b"},
		{"/txt/1/c.txt", "c"},
		{"/txt/welcome.txt", "Welcome to the page!"},
	}
	for i := range tests {
		resp, err := http.Get(ts.URL + tests[i][0])
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.EqualFold(tests[i][1], string(data)) {
			t.Errorf("%d expected %s got %s", i, tests[i][1], string(data))
		}
	}
	resp, err := http.Get(ts.URL + "/txt/hi.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
}

// 404 page not found

func TestCacheFSGZIP(t *testing.T) {
	var mf MemoryFile
	welcome := "Welcome to the page!"
	r := &WRoute{Mux: http.NewServeMux()}
	err := r.CacheFileGZIP(flate.DefaultCompression, "txt\\welcome.txt", &mf)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	client := &http.Client{}
	req, _ := http.NewRequest("GET", ts.URL+"/txt/welcome.txt", nil)
	req.Header.Add("Accept-Encoding", "gzip")
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(res.Header.Get("ETag"), "f7b3b7e0f288ecbc7d805c12d782cb88") {
		t.Errorf("got %s | expected f7b3b7e0f288ecbc7d805c12d782cb88", res.Header.Get("ETag"))
	}
	if !strings.EqualFold(res.Header.Get("Content-Encoding"), "gzip") {
		t.Errorf("got %s | expected gzip", res.Header.Get("Content-Encoding"))
	}
	if !strings.EqualFold(res.Header.Get("Vary"), "Accept-Encoding") {
		t.Errorf("got %s | expected Accept-Encoding", res.Header.Get("Vary"))
	}
	gr, err := gzip.NewReader(res.Body)
	defer res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	defer gr.Close()
	unzipped := make([]byte, 64)
	n, err := gr.Read(unzipped)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(unzipped[:n], []byte(welcome)) {
		t.Errorf("got %s | expected %s", string(unzipped[:n]), welcome)
	}
}

func TestStoreTemplate(t *testing.T) {
	var mf MemoryFile
	welcome := "Welcome to the home page!"
	load := func() ([]byte, error) {
		tl, err := template.ParseFiles("file.tmpl")
		if err != nil {
			return nil, fmt.Errorf("缓存模板(step1)失败: %w", err)
		}
		buf := bytes.NewBuffer(make([]byte, 0, 64))
		err = tl.Execute(buf, welcome)
		if err != nil {
			return nil, fmt.Errorf("缓存模板(step2)失败: %w", err)
		}
		return buf.Bytes(), nil
	}
	mf.StoreFileBuffer("a", load)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var err error
		_, err = mf.WriteTo("a", w)
		if err != nil {
			t.Fatal("err: ", err.Error())
		}
	}))
	defer ts.Close()
	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, []byte(welcome)) {
		t.Errorf("got %s | expected %s", string(data), welcome)
	}
}
func TestStoreTemplateGZIP(t *testing.T) {
	var mf MemoryFile
	welcome := "Welcome to the home page!"
	load := func() ([]byte, error) {
		t, err := template.ParseFiles("file.tmpl")
		if err != nil {
			return nil, fmt.Errorf("缓存GZIP压缩模板(step1)失败: %w", err)
		}
		var buf1, buf2 bytes.Buffer
		err = t.Execute(&buf1, welcome)
		if err != nil {
			return nil, fmt.Errorf("缓存GZIP压缩模板(step2)失败: %w", err)
		}
		gz, err := gzip.NewWriterLevel(&buf2, gzip.DefaultCompression)
		if err != nil {
			return nil, fmt.Errorf("缓存GZIP压缩模板(step3)失败: %w", err)
		}
		defer gz.Close()
		_, err = gz.Write(buf1.Bytes())
		if err != nil {
			return nil, fmt.Errorf("缓存GZIP压缩模板(step4)失败: %w", err)
		}
		err = gz.Flush()
		if err != nil {
			return nil, fmt.Errorf("缓存GZIP压缩模板(step5)失败: %w", err)
		}
		return buf2.Bytes(), nil
	}
	mf.StoreFileBuffer("a", load)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
			t.Fatal("非gzip请求")
		}
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		_, err := mf.WriteTo("a", w)
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer ts.Close()
	client := &http.Client{}
	req, _ := http.NewRequest("GET", ts.URL, nil)
	req.Header.Add("Accept-Encoding", "gzip")
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	gz, err := gzip.NewReader(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	unzipped := make([]byte, 64)
	n, err := gz.Read(unzipped)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(unzipped[:n], []byte(welcome)) {
		t.Errorf("got %s | expected %s", string(unzipped[:n]), welcome)
	}
}
