package whttp

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClickjackingMiddleware(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	r := &WRoute{Mux: http.NewServeMux()}
	r.Use(ClickjackingMiddleware())
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
