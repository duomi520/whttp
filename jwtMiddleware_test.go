package whttp

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestJWTMiddleware(t *testing.T) {
	jwt := JWT{TokenSigningKey: []byte("TokenSigningKey"), TokenExpires: time.Duration(time.Second)}
	r := NewRoute(nil)
	r.Mux = http.NewServeMux()
	fn := func(c *HTTPContext) {
		id, _ := c.Get("id")
		if id.(float64) != 1920 {
			t.Fatal("不等于1920")
		}
		if err := jwt.RefreshToken(c); err != nil {
			t.Fatal(err.Error())
		}
		c.String(http.StatusOK, "Hi")
	}
	r.POST("/", jwt.JWTMiddleware("id"), fn)
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
	claims := map[string]any{"id": 1920}
	token, err := jwt.CreateToken(claims)
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

// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTMxODA3ODUsImlhdCI6MTc1MzE4MDc4NCwiaWQiOjE5MjB9.LiDiFAd6I-0fmk2HuZ5kw1BJcs6KCO8wRzlBs74yoBQ
