package whttp

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestETagMiddleware(t *testing.T) {
	var etag sync.Map
	slog.SetLogLoggerLevel(slog.LevelDebug)
	fn := func(c *HTTPContext) { c.String(200, "Hi") }
	r := NewRoute(nil)
	r.Mux = http.NewServeMux()
	r.GET("/e", LoggerMiddleware(), ETagMiddleware(&etag), fn)
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	res, err := http.Get(ts.URL + "/e")
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	ETag := res.Header.Get("ETag")
	if !strings.EqualFold(ETag, "c1a5298f939e87e8f962a5edfc206918") {
		t.Error("invalid ETag:", ETag)
	}
	fmt.Println(res.StatusCode, string(body))
	client := &http.Client{}
	req, err := http.NewRequest("GET", ts.URL+"/e", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("If-None-Match", ETag)
	var sc int
	for range 3 {
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		sc = resp.StatusCode
	}
	fmt.Println(sc, string(body))
}

/*
2025/08/04 12:35:34 DEBUG | 0s            | 127.0.0.1:54626 | 200 | GET     | /e                                       |       2 bytes
200 Hi
2025/08/04 12:35:34 DEBUG | 0s            | 127.0.0.1:54626 | 304 | GET     | /e                                       |       0 bytes
2025/08/04 12:35:34 DEBUG | 0s            | 127.0.0.1:54626 | 304 | GET     | /e                                       |       0 bytes
2025/08/04 12:35:34 DEBUG | 0s            | 127.0.0.1:54626 | 304 | GET     | /e                                       |       0 bytes
304
*/
