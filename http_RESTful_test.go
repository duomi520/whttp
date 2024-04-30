package whttp

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestFormValue(t *testing.T) {
	r := &WRoute{mux: http.NewServeMux()}
	fn := func(c *HTTPContext) {
		if (strings.Compare(c.Request.FormValue("name"), "linda") == 0) && (strings.Compare(c.Request.FormValue("mobile"), "xxxxxxxx") == 0) {
			c.String(http.StatusOK, "OK")
		} else {
			c.String(http.StatusOK, "NG")
		}
	}
	r.POST("/", fn)
	ts := httptest.NewServer(r.mux)
	defer ts.Close()
	res, err := http.Post(ts.URL, "application/x-www-form-urlencoded",
		strings.NewReader("name=linda&mobile=xxxxxxxx"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, []byte("OK")) {
		t.Errorf("got %s | expected OK", string(data))
	}

}
func TestParams(t *testing.T) {
	r := &WRoute{mux: http.NewServeMux()}
	fn := func(c *HTTPContext) {
		if (strings.Compare(c.Request.PathValue("name"), "linda") == 0) && (strings.Compare(c.Request.PathValue("mobile"), "xxxxxxxx") == 0) {
			c.String(http.StatusOK, "OK")
		} else {
			c.String(http.StatusOK, "NG")
		}
	}
	r.GET("/user/{name}/mobile/{mobile}", fn)
	ts := httptest.NewServer(r.mux)
	defer ts.Close()
	res, err := http.Get(ts.URL + "/user/linda/mobile/xxxxxxxx")
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, []byte("OK")) {
		t.Errorf("got %s | expected OK", string(data))
	}

}

func TestMiddleware(t *testing.T) {
	signature := ""
	r := &WRoute{mux: http.NewServeMux()}
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
	fn := func(c *HTTPContext) {
		signature += "<->"
	}
	r.POST("/", MiddlewareA(), MiddlewareB(), MiddlewareC(), fn)
	ts := httptest.NewServer(r.mux)
	defer ts.Close()
	_, err := http.Post(ts.URL, "application/x-www-form-urlencoded",
		strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(signature, "A1B1C1<->C2B2A2") {
		t.Errorf("got %s | expected A1B1C1<->C2B2A2", signature)
	}
}

func TestMethod(t *testing.T) {
	r := &WRoute{mux: http.NewServeMux()}
	r.logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format(time.DateTime))
				}
			}
			return a
		},
	}))
	fn := func(c *HTTPContext) {}
	r.GET("/", fn)
	ts := httptest.NewServer(r.mux)
	defer ts.Close()
	_, err := http.Post(ts.URL, "application/x-www-form-urlencoded", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
}

/*
time="2024-05-01 00:15:08" level=ERROR msg="not a POST request"
*/
