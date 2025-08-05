package whttp

import (
	"bytes"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
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
	for range 10 {
		_, err := http.Post(ts.URL, "application/x-www-form-urlencoded",
			strings.NewReader(""))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.EqualFold(signature, "A1B1C1<->C2B2A2") {
			t.Errorf("got %s | expected A1B1C1<->C2B2A2", signature)
		}
		signature = ""
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
	for range 5 {
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
}

func TestRender(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
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
	r.GET("/", LoggerMiddleware(), fn)
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	for range 2 {
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
}

// 2025/08/04 12:34:26 DEBUG | 0s            | 127.0.0.1:54554 | 200 | GET     | /                                        |       8 bytes
// 2025/08/04 12:34:26 DEBUG | 0s            | 127.0.0.1:54554 | 200 | GET     | /                                        |       8 bytes

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
	tests := [][3]string{
		{"/", "txt\\welcome.txt", "Welcome to the page!"},
		{"/c", "txt\\1\\c.txt", "c"},
		{"/d", "txt\\d.txt", "404 page not found"},
		{"/e", "txt\\1\\《洛神赋》.txt", "黄初三年"},
	}
	r := NewRoute(nil)
	r.Mux = http.NewServeMux()
	for i := range tests {
		r.Static(tests[i][0], tests[i][1])
	}
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	for i := range tests {
		res, err := http.Get(ts.URL + "/" + tests[i][0][1:])
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(res.Body)
		defer res.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if i == 3 {
			data = data[:12]
		}
		if !bytes.Equal(data, []byte(tests[i][2])) {
			t.Errorf("got %s | expected %s", string(data), tests[i][2])
		}
	}
}

// 2025/07/28 11:56:16 ERROR File error="open txt\\d.txt: The system cannot find the file specified."
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
		{"/txt/1/《洛神赋》.txt", "黄初三年"},
	}
	for i := range tests {
		filePath, fileName := filepath.Split(tests[i][0])
		escapeUrl := filePath + url.QueryEscape(fileName)
		resp, err := http.Get(ts.URL + escapeUrl)
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if i == 6 {
			data = data[:12]
		}
		if !strings.EqualFold(tests[i][1], string(data)) {
			t.Errorf("%d expected %s got %s", i, tests[i][1], string(data))
		}
	}
}
