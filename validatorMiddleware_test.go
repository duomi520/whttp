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
		if !strings.EqualFold(c.Request.PathValue("number"), "777") {
			t.Fatal(c.Request.PathValue("number"))
		}
	}
	r := NewRoute(validator.New(), nil)
	r.Mux = http.NewServeMux()
	r.POST("/number/{number}", ValidatorMiddleware("number:numeric"), fn)
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	resp, err := http.Post(ts.URL+"/number/777", "application/x-www-form-urlencoded",
		strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatal(string(data))
	}
}
