package whttp

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestClickjacking(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	r := NewRoute(validator.New(), nil)
	r.Mux = http.NewServeMux()
	r.Use(HeaderMiddleware(map[string]string{"frame-ancestors": "none", "X-Frame-Optoins": "DENY"}))
	fn := func(c *HTTPContext) {
		c.String(200, "Hi")
	}
	r.GET("/", fn)
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(res.Header.Get("frame-ancestors"), "none") {
		t.Errorf("got %s | expected none", res.Header.Get("frame-ancestors"))
	}
	if !strings.EqualFold(res.Header.Get("X-Frame-Optoins"), "DENY") {
		t.Errorf("got %s | expected DENY", res.Header.Get("X-Frame-Optoins"))
	}
}
