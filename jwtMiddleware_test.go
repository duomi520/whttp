package whttp

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
)

func TestJWTMiddleware(t *testing.T) {
	j := JWT{TokenSigningKey: []byte("TokenSigningKey"), TokenExpires: time.Duration(time.Second)}
	r := NewRoute(validator.New(), nil)
	r.Mux = http.NewServeMux()
	fn := func(c *HTTPContext) {
		id, _ := c.Get("id")
		if id.(float64) != 1920 {
			t.Fatal("不等于1920")
		}
		c.Set("SecretKey", []string{"id"})
		c.Set("SecretObj", []any{1920})
		c.String(http.StatusOK, "Hi")
	}
	r.POST("/", j.JWTMiddleware("id"), fn)
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
	resp2, err := (&http.Client{}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	data2, err := io.ReadAll(resp2.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(string(data2), "Hi") {
		t.Fatalf("expected %v got %v", "Hi", string(data2))
	}
	// 读取所有响应头部
	header := resp2.Header
	fmt.Println(header.Get("Authorization"))
}

// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTkxMjc3NjYsImlhdCI6MTcxOTEyNzc2NSwiaWQiOjE5MjB9.N1XbLFECK9N8ZtdN7QS-I_M5EY8alOOpMpWED0teNMs
