package whttp

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var mf MemoryFile

func TestCacheTemplate(t *testing.T) {
	welcome := "Welcome to the home page!"
	mf.CacheTemplate("memoryFile.tmpl", "a", welcome)
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

func TestCacheTemplateGZIP(t *testing.T) {
	welcome := "Welcome to the home page!"
	mf.CacheTemplateGZIP(gzip.DefaultCompression, "memoryFile.tmpl", "a", welcome)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, err := mf.WriteGZIP("a", w, req)
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
	data, err := io.ReadAll(gz)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, []byte(welcome)) {
		t.Errorf("got %s | expected %s", string(data), welcome)
	}
}

func TestCacheFile(t *testing.T) {
	welcome := "Welcome to the page!"
	r := &WRoute{Mux: http.NewServeMux()}
	err := r.CacheFile("/txt/welcome.txt", &mf)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	res, err := http.Get(ts.URL + "/txt/welcome.txt")
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, []byte(welcome)) {
		t.Errorf("got %s | expected %s", string(data), welcome)
	}
}

func TestCacheFS(t *testing.T) {
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
