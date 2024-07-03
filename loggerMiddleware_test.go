package whttp

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestLoggerMiddleware(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	fn := func(c *HTTPContext) {}
	r := NewRoute(validator.New(), nil)
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
2024/05/19 13:26:54 DEBUG |            0s | 127.0.0.1:49389 |     0 |    POST | /p | 0
2024/05/19 13:26:54 DEBUG |            0s | 127.0.0.1:49389 |     0 |     GET | /g | 0
*/
