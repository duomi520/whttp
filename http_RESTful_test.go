package whttp

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/duomi520/utils"
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
	r.logger, _ = utils.NewWLogger(utils.InfoLevel, "")
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
[Error] 2024-04-27 13:19:02 This is not a POST request.
*/
