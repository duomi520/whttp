package whttp

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/csrf"
)

func getCSRF(c *HTTPContext) {
	// 获取token值
	token := csrf.Token(c.Request)
	// 将token写入到header中
	c.Writer.Header().Set("X-CSRF-Token", token)
	// 业务代码
	fmt.Fprintln(c.Writer, "hello")
}

func postCSRF(c *HTTPContext) {
	token := csrf.Token(c.Request)
	c.Writer.Header().Set("X-CSRF-Token", token)
	// 业务代码
	fmt.Fprintln(c.Writer, "<1>")
}

func TestCSRFMiddleware(t *testing.T) {
	r := &WRoute{Mux: http.NewServeMux()}
	csrfMiddleware := csrf.Protect([]byte("32-byte-long-auth-key"))
	r.GET("/a", getCSRF)
	r.POST("/c", postCSRF)
	//csrfMiddleware 默认只对POST验证
	ts := httptest.NewServer(csrfMiddleware(r.Mux))
	defer ts.Close()
	//先get
	res, err := http.Get(ts.URL + "/a")
	if err != nil {
		t.Fatal(err)
	}
	token := res.Header.Get("X-CSRF-Token")
	if len(token) == 0 {
		t.Fatal("token is nil")
	}
	cookies := res.Cookies()
	if len(cookies[0].String()) == 0 {
		t.Fatal("cookies is nil")
	}
	//不带token
	req, err := http.NewRequest("POST", ts.URL+"/c", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatal(resp.Status)
	}
	//fmt.Println(resp)
	fmt.Println("1 ", resp.Header.Get("X-CSRF-Token"))
	//带token
	req.Header.Set("Cookie", cookies[0].String())
	req.Header.Set("X-CSRF-Token", token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err = (&http.Client{}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatal(resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("2 ", string(data))
	fmt.Println("3 ", resp.Header.Get("X-CSRF-Token"))
	//fmt.Println(resp)
}

/*
1
2  <1>

3  nXS7COmuLabTPdqddS3ujNU+1zWJ2J+eQWNgA76xE0VPgcrp2/ua4rJ5GVwpCgPXMRQRy2KVlRjesPXXQZTtnQ==
*/

// https://studygolang.com/articles/35927
