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
	fn := func(c *HTTPContext) {}
	r := &WRoute{mux: http.NewServeMux()}
	r.POST("/p", LoggerMiddleware(), fn)
	r.GET("/g", LoggerMiddleware(), fn)
	ts := httptest.NewServer(r.mux)
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
2024/04/30 22:31:16 DEBUG |            0s | 127.0.0.1:54122 |     0 |    POST | /p |
2024/04/30 22:31:16 DEBUG |            0s | 127.0.0.1:54122 |     0 |     GET | /g |
*/
