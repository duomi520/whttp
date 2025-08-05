package whttp

import (
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type SyncMapCache struct {
	sync.Map
}

func (sc *SyncMapCache) Set(k string, v []byte) {
	b := make([]byte, len(v))
	copy(b, v)
	sc.Store(k, b)
}
func (sc *SyncMapCache) HasGet(dst []byte, k string) ([]byte, bool) {
	v, ok := sc.Load(k)
	if ok {
		return v.([]byte), ok
	}
	return nil, ok
}
func (sc *SyncMapCache) Del(k string) {
	sc.Delete(k)
}

func TestCacheMiddleware(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	sum := 0
	fn := func(c *HTTPContext) {
		sum++
		c.String(200, "世间安得两全法，不负如来不负卿。")
	}
	r := NewRoute(nil)
	r.Mux = http.NewServeMux()
	var cache SyncMapCache
	r.GET("/c",
		LoggerMiddleware(),
		CacheMiddleware(&cache, map[string]string{"Content-Type": "text/html"}),
		fn)
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	for range 3 {
		getHi(t, ts.URL)
	}
	if sum != 1 {
		t.Fatal("缓存失败")
	}
}

func getHi(t *testing.T, url string) {
	res, err := http.Get(url + "/c")
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if !strings.EqualFold(string(body), "世间安得两全法，不负如来不负卿。") {
		t.Fatal("接收不到:世间安得两全法，不负如来不负卿。")
	}
}

/*
2025/08/05 15:02:22 DEBUG | 0s            | 127.0.0.1:54619 | 200 | GET     | /c                                       |      48 bytes
2025/08/05 15:02:22 DEBUG | 0s            | 127.0.0.1:54619 | 200 | GET     | /c                                       |      48 bytes
2025/08/05 15:02:22 DEBUG | 0s            | 127.0.0.1:54619 | 200 | GET     | /c                                       |      48 bytes
*/
func TestCacheGZIPFile(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	r := NewRoute(nil)
	r.Mux = http.NewServeMux()
	var cache SyncMapCache
	g := []func(*HTTPContext){
		LoggerMiddleware(),
		CacheMiddleware(&cache, map[string]string{
			"Content-Type":     "text/html; charset=utf-8",
			"Content-Encoding": "gzip",
			"Vary":             "Accept-Encoding",
		}),
		GZIPMiddleware(gzip.DefaultCompression),
	}
	fn := func(c *HTTPContext) {
		c.File("./txt/1/《滕王阁序》.txt")
	}
	r.GET("/t", append(g, fn)...)
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	for range 2 {
		data := getResp(t, ts.URL+"/t", nil)
		if !strings.EqualFold(string(data[:27]), "豫章故郡，洪都新府") {
			t.Fatalf("got %s | expected 豫章故郡，洪都新府", string(data[:27]))
		}
	}
}
/*
2025/08/05 15:36:38 DEBUG | 87.6818ms     | 127.0.0.1:58416 | 200 | GET     | /t                                       |    1904 bytes
Content-Type : [text/plain; charset=utf-8]
Last-Modified : [Tue, 05 Aug 2025 07:10:38 GMT]
Vary : [Accept-Encoding]
Date : [Tue, 05 Aug 2025 07:36:38 GMT]
2025/08/05 15:36:38 DEBUG | 0s            | 127.0.0.1:58416 | 200 | GET     | /t                                       |    1904 bytes
Content-Type : [text/html; charset=utf-8]
Vary : [Accept-Encoding]
Date : [Tue, 05 Aug 2025 07:36:38 GMT]
*/