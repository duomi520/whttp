package whttp

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator"
)

func TestValidatorMiddleware(t *testing.T) {
	fn := func(c *HTTPContext) {
		if !strings.EqualFold(c.Request.PathValue("a"), "777") {
			t.Fatal(c.Request.PathValue("a"))
		}
	}
	r := NewRoute(validator.New(), nil)
	r.Mux = http.NewServeMux()
	r.POST("/a/{a}", ValidatorMiddleware("a:numeric", "b:lt=5"), fn)
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	resp, err := http.Post(ts.URL+"/a/777", "application/x-www-form-urlencoded",
		strings.NewReader("b=667&d=hi"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatal(string(data))
	}
}
