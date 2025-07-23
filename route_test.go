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
)

func TestFormValue(t *testing.T) {
	r := NewRoute(nil)
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
	r := NewRoute(nil)
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
	r := NewRoute(nil)
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
	r := NewRoute(nil)
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
time="2025-07-23 16:56:37" level=ERROR msg="not a POST request" url=/
*/

func TestFile(t *testing.T) {
	r := NewRoute(nil)
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
	r := NewRoute(nil)
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
	r := NewRoute(nil)
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

// 2025/07/23 09:53:26 DEBUG | 40.6µs        | 127.0.0.1:63189 | 200 | GET     | /                                        |       2 bytes
type testUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

func TestBindJSON(t *testing.T) {
	r := NewRoute(nil)
	r.Mux = http.NewServeMux()
	r.Use(LoggerMiddleware())
	fn := func(c *HTTPContext) {
		var user testUser
		if err := c.BindJSON(&user); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		c.String(200, user.Username+":"+user.Email)
	}
	r.GET("/", fn)
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	req, err := http.NewRequest("GET", ts.URL, strings.NewReader(`{"username":"runoob", "email": "xxx@163.com"}`))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, []byte("runoob:xxx@163.com")) {
		t.Errorf("got %s | expected runoob:xxx@163.com", string(data))
	}
}

func TestStatic(t *testing.T) {
	r := NewRoute(nil)
	r.Mux = http.NewServeMux()
	r.Static("/", "txt\\welcome.txt")
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, []byte("Welcome to the page!")) {
		t.Errorf("got %s | expected %s", string(data), "Welcome to the page!")
	}
}
func TestStaticFS(t *testing.T) {
	r := NewRoute(nil)
	r.Mux = http.NewServeMux()
	r.StaticFS("txt")
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	tests := [][2]string{
		{"/txt/a.txt", "a"},
		{"/txt/b.txt", "b"},
		{"/txt/1/c.txt", "c"},
		{"/txt/welcome.txt", "Welcome to the page!"},
		{"/file.tmp", "404 page not found\n"},
		{"/file.tmpl", "404 page not found\n"},
	}
	for i := range tests {
		resp, err := http.Get(ts.URL + tests[i][0])
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.EqualFold(tests[i][1], string(data)) {
			t.Errorf("%d expected %s got %s", i, tests[i][1], string(data))
		}
	}
}
