package whttp

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggerMiddleware(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	fn := func(c *HTTPContext) { c.String(200, "Hi") }
	r := NewRoute(nil)
	r.Mux = http.NewServeMux()
	r.POST("/p", LoggerMiddleware(), fn)
	r.GET("/g", LoggerMiddleware(), fn)
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	_, err := http.Post(ts.URL+"/p", "application/x-www-form-urlencoded",
		strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	_, err = http.Get(ts.URL + "/g")
	if err != nil {
		t.Fatal(err)
	}
}

/*
2025/07/21 20:51:03 DEBUG | 0s            | 127.0.0.1:53422 | 200 | POST    | /p                                       |       2 bytes
2025/07/21 20:51:03 DEBUG | 0s            | 127.0.0.1:53423 | 200 | GET     | /g                                       |       2 bytes
*/
