package whttp

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

var showRespHeader bool = true

func getResp(t *testing.T, url string, header map[string]string) []byte {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}
	//client.Do 默认自动解压，当设置"Accept-Encoding"时，body需手动解压
	//req.Header.Add("Accept-Encoding", "gzip")
	//req.Header.Set("Accept-Encoding", "")
	for k, v := range header {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if showRespHeader {
		for k, v := range resp.Header {
			fmt.Println(k, ":", v)
		}
	}
	return data
}

func TestGZIPMiddleware(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	fn := func(c *HTTPContext) { c.String(200, "未压缩内容") }
	r := NewRoute(nil)
	r.Mux = http.NewServeMux()
	r.GET("/g", GZIPMiddleware(gzip.DefaultCompression), fn)
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	for range 2 {
		data := getResp(t, ts.URL+"/g", nil)
		if !bytes.Equal(data, []byte("未压缩内容")) {
			t.Fatal(string(data))
		}
	}
}

/*
Vary : [Accept-Encoding]
Date : [Sun, 03 Aug 2025 11:26:51 GMT]
Vary : [Accept-Encoding]
Date : [Sun, 03 Aug 2025 11:26:51 GMT]
*/

func TestGZIPFile(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	r := NewRoute(nil)
	r.Mux = http.NewServeMux()
	fn := func(c *HTTPContext) {
		c.File("./txt/a.txt")
	}
	r.GET("/", GZIPMiddleware(gzip.DefaultCompression), fn)
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	for range 2 {
		data := getResp(t, ts.URL, nil)
		if !bytes.Equal(data, []byte("a")) {
			t.Errorf("got %s | expected a", string(data))
		}
	}
}

/*
Content-Type : [text/plain; charset=utf-8]
Last-Modified : [Mon, 28 Jul 2025 02:05:06 GMT]
Vary : [Accept-Encoding]
Date : [Sun, 03 Aug 2025 11:27:08 GMT]
Date : [Sun, 03 Aug 2025 11:27:08 GMT]
Content-Type : [text/plain; charset=utf-8]
Last-Modified : [Mon, 28 Jul 2025 02:05:06 GMT]
Vary : [Accept-Encoding]
*/

func TestETagGZIPFile(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	r := NewRoute(nil)
	r.Mux = http.NewServeMux()
	var etag sync.Map
	g := []func(*HTTPContext){LoggerMiddleware(), ETagMiddleware(&etag), GZIPMiddleware(gzip.DefaultCompression)}
	fn := func(c *HTTPContext) {
		c.File("./txt/1/《洛神赋》.txt")
	}
	r.GET("/", append(g, fn)...)
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	data := getResp(t, ts.URL, nil)
	if !strings.EqualFold(string(data[:24]), "黄初三年，余朝京") {
		t.Errorf("got %s | expected 黄初三年，余朝京", string(data[:24]))
	}
	h1 := map[string]string{"If-None-Match": "24d58daae35a37265a5e40757dc0c686"}
	data = getResp(t, ts.URL, h1)
	if len(data) > 0 {
		t.Error("data 不为空")
	}
	h2 := map[string]string{"If-Modified-Since": "Sun, 20 Jul 2025 02:05:06 GMT"}
	data = getResp(t, ts.URL, h2)
	if !strings.EqualFold(string(data[:24]), "黄初三年，余朝京") {
		t.Errorf("got %s | expected 黄初三年，余朝京", string(data[:24]))
	}
	h3 := map[string]string{"If-Modified-Since": "Sun, 10 Aug 2025 02:05:06 GMT"}
	data = getResp(t, ts.URL, h3)
	if len(data) > 0 {
		t.Error("data 不为空")
	}
}

/*
2025/08/04 12:37:16 DEBUG | 55.9222ms     | 127.0.0.1:54750 | 200 | GET     | /                                        |    2079 bytes
Etag : [24d58daae35a37265a5e40757dc0c686]
Last-Modified : [Mon, 28 Jul 2025 02:05:06 GMT]
Vary : [Accept-Encoding]
Date : [Mon, 04 Aug 2025 04:37:16 GMT]
Content-Type : [text/plain; charset=utf-8]
2025/08/04 12:37:16 DEBUG | 0s            | 127.0.0.1:54750 | 304 | GET     | /                                        |       0 bytes
Date : [Mon, 04 Aug 2025 04:37:16 GMT]
2025/08/04 12:37:16 DEBUG | 1.0539ms      | 127.0.0.1:54750 | 200 | GET     | /                                        |    2079 bytes
Content-Type : [text/plain; charset=utf-8]
Etag : [24d58daae35a37265a5e40757dc0c686]
Last-Modified : [Mon, 28 Jul 2025 02:05:06 GMT]
Vary : [Accept-Encoding]
Date : [Mon, 04 Aug 2025 04:37:16 GMT]
2025/08/04 12:37:16 WARN write error="http: request method or response status code does not allow body"
2025/08/04 12:37:16 DEBUG | 0s            | 127.0.0.1:54750 | 304 | GET     | /                                        |      23 bytes
Content-Encoding : [gzip]
Last-Modified : [Mon, 28 Jul 2025 02:05:06 GMT]
Vary : [Accept-Encoding]
Date : [Mon, 04 Aug 2025 04:37:16 GMT]
*/

func TestGZIPRande(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	tl, err := template.ParseFiles("file.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	r := NewRoute(nil)
	r.Mux = http.NewServeMux()
	r.SetRenderer(tl)
	g := []func(*HTTPContext){LoggerMiddleware(), GZIPMiddleware(gzip.DefaultCompression)}
	fn := func(c *HTTPContext) {
		c.Render(http.StatusOK, "file.tmpl", "人生若只如初见")
	}
	r.GET("/", append(g, fn)...)
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	data := getResp(t, ts.URL, nil)
	if !strings.EqualFold(string(data), "人生若只如初见") {
		t.Errorf("got %s | expected 人生若只如初见", string(data))
	}
}

/*
2025/08/04 14:38:55 DEBUG | 516.3µs       | 127.0.0.1:61552 | 200 | GET     | /                                        |      47 bytes
Date : [Mon, 04 Aug 2025 06:38:55 GMT]
Content-Type : [text/html]
Vary : [Accept-Encoding]
*/
