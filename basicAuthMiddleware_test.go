package whttp

import (
	"bytes"
	"encoding/base64"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBasicAuthMiddlewareMiddleware(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	r := NewRoute(nil)
	r.Mux = http.NewServeMux()
	v := func(c *HTTPContext, u, p string) bool {
		if u == "joe" && p == "secret" {
			return true
		}
		return false
	}
	fn := func(c *HTTPContext) {
		c.String(200, "Hi")
	}
	r.GET("/", BasicAuthMiddleware(v), fn)
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	client := &http.Client{}
	req, _ := http.NewRequest("GET", ts.URL, nil)
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(res.Status, "401 Unauthorized") {
		t.Errorf("got %s | expected 401 Unauthorized", res.Status)
	}
	auth := "Basic " + base64.StdEncoding.EncodeToString([]byte("joe:secret"))
	req.Header.Add("Authorization", auth)
	res, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, []byte("Hi")) {
		t.Errorf("got %s | expected Hi", string(data))
	}
}
