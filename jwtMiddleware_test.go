package whttp

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt/v4"
)

func TestJWTMiddleware(t *testing.T) {
	j := JWT{TokenSigningKey: []byte("TokenSigningKey"), TokenExpires: time.Duration(time.Second)}
	r := NewRoute(validator.New(), nil)
	r.Mux = http.NewServeMux()
	fn := func(c *HTTPContext) {
		claims, ok := c.Get("claims")
		if !ok {
			t.Fatal("claims缺失")
		}
		if claims.(jwt.MapClaims)["id"].(float64) != 1920 {
			t.Fatal("不等于1920")
		}
	}
	r.POST("/", j.JWTMiddleware(), fn)
	ts := httptest.NewServer(r.Mux)
	defer ts.Close()
	//不带token
	req, err := http.NewRequest("POST", ts.URL, nil)
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
	if !strings.EqualFold(string(data), "缺少令牌") {
		t.Fatalf("expected %v got %v", "缺少令牌", string(data))
	}
	//带token
	key := []string{"id"}
	obj := []any{1920}
	token, err := j.CreateToken(key, obj)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", token)
	_, err = (&http.Client{}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
}
