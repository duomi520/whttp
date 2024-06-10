package whttp

import (
	"bytes"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator"
)

func TestFormValue(t *testing.T) {
	r := NewRoute(validator.New(), nil)
	r.Mux = http.NewServeMux()
	fn := func(c *HTTPContext) {
		if (strings.Compare(c.Request.FormValue("name"), "linda") == 0) && (strings.Compare(c.Request.FormValue("mobile"), "xxxxxxxx") == 0) {
			c.String(http.StatusOK, "OK")
		} else {
			c.String(http.StatusOK, "NG")
		}
	}
	r.POST("/", fn)
	ts := httptest.NewServer(r.Mux)
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
	r := NewRoute(validator.New(), nil)
	r.Mux = http.NewServeMux()
	fn := func(c *HTTPContext) {
		if (strings.Compare(c.Request.PathValue("name"), "linda") == 0) && (strings.Compare(c.Request.PathValue("mobile"), "xxxxxxxx") == 0) {
			c.String(http.StatusOK, "OK")
		} else {
			c.String(http.StatusOK, "NG")
		}
	}
	r.GET("/user/{name}/mobile/{mobile}", fn)
	ts := httptest.NewServer(r.Mux)
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
	r := NewRoute(validator.New(), nil)
	r.Mux = http.NewServeMux()
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
	g := []func(*HTTPContext){MiddlewareA(), MiddlewareB(), MiddlewareC()}
	r.POST("/", append(g, fn)...)
	//r.POST("/", MiddlewareA(), MiddlewareB(), MiddlewareC(), fn)
	ts := httptest.NewServer(r.Mux)
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
	r := NewRoute(validator.New(), nil)
	r.Mux = http.NewServeMux()
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
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	res, err := http.Post(ts.URL, "application/x-www-form-urlencoded", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, []byte("method not allowed")) {
		t.Errorf("got %s | expected method not allowed", string(data))
	}
}

/*
time="2024-05-01 00:15:08" level=ERROR msg="not a POST request"
*/

func TestFile(t *testing.T) {
	r := NewRoute(validator.New(), nil)
	r.Mux = http.NewServeMux()
	fn := func(c *HTTPContext) {
		c.File("txt/a.txt")
	}
	r.GET("/", fn)
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, []byte("a")) {
		t.Errorf("got %s | expected a", string(data))
	}
}

func TestRender(t *testing.T) {
	tl, err := template.ParseFiles("file.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	r := NewRoute(validator.New(), nil)
	r.Mux = http.NewServeMux()
	r.SetRenderer(tl)
	fn := func(c *HTTPContext) {
		c.Render(http.StatusOK, "file.tmpl", "6月7日")
	}
	r.GET("/", fn)
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, []byte("6月7日")) {
		t.Errorf("got %s | expected 6月7日", string(data))
	}
}

func TestUse(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	r := NewRoute(validator.New(), nil)
	r.Mux = http.NewServeMux()
	r.Use(LoggerMiddleware())
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
	data, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, []byte("Hi")) {
		t.Errorf("got %s | expected Hi", string(data))
	}
}

// 2024/05/23 19:44:19 DEBUG |       514.2µs | 127.0.0.1:51959 |   200 |     GET | / | 2
