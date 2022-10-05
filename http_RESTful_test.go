package whttp

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestWarp(t *testing.T) {
	r := &WRoute{router: mux.NewRouter()}
	fn := func(c *HTTPContext) {
		if (strings.Compare(c.Params("name"), "linda") == 0) && (strings.Compare(c.Params("mobile"), "xxxxxxxx") == 0) {
			c.String(http.StatusOK, "OK")
		} else {
			c.String(http.StatusOK, "NG")
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(r.Warp(nil, fn)))
	defer ts.Close()
	res, err := http.Post(ts.URL, "application/x-www-form-urlencoded",
		strings.NewReader("name=linda&mobile=xxxxxxxx"))
	if err != nil {
		log.Fatal(err)
	}
	greeting, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	if !bytes.Equal(greeting, []byte("OK")) {
		t.Errorf("got %s | expected ok", string(greeting))
	}
}

func TestMiddleware(t *testing.T) {
	signature := ""
	r := &WRoute{router: mux.NewRouter()}
	MiddlewareA := func() func(*HTTPContext) {
		return func(c *HTTPContext) {
			signature += "A1"
			c.Next()
			signature += "A2"
		}
	}
	MiddlewareB := func() func(*HTTPContext) {
		return func(c *HTTPContext) {
			signature += "B1"
			c.Next()
			signature += "B2"
		}
	}
	MiddlewareC := func() func(*HTTPContext) {
		return func(c *HTTPContext) {
			signature += "C1"
			c.Next()
			signature += "C2"
		}
	}
	group := HTTPMiddleware(MiddlewareA(), MiddlewareB(), MiddlewareC())
	fn := func(c *HTTPContext) {
		signature += "<->"
	}
	ts := httptest.NewServer(http.HandlerFunc(r.Warp(group, fn)))
	defer ts.Close()
	_, err := http.Post(ts.URL, "application/x-www-form-urlencoded",
		strings.NewReader(""))
	if err != nil {
		log.Fatal(err)
	}
	if !strings.EqualFold(signature, "A1B1C1<->C2B2A2") {
		t.Errorf("got %s | expected A1B1C1<->C2B2A2", signature)
	}
}
