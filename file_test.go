package whttp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var mf MemoryFile

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
	res.Body.Close()
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
	data, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(res.Header.Get("ETag"), "f615cf810b8c9723d0a836dbd9df8648") {
		t.Errorf("got %s | expected f615cf810b8c9723d0a836dbd9df8648", res.Header.Get("ETag"))
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
