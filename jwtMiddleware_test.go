package whttp

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
)

func TestJWT(t *testing.T) {
	jwt := JWT{TokenSigningKey: []byte("TokenSigningKey"), TokenExpires: time.Duration(time.Second)}
	group := HTTPMiddleware(jwt.JWTMiddleware())
	r := &WRoute{router: httprouter.New()}
	fn := func(c *HTTPContext) {
		tokenString := c.Request.Header["Authorization"][0]
		data, err := jwt.TokenParse(tokenString)
		if err != nil {
			t.Fatal(err)
		}
		if data.(float64) != 1920 {
			t.Fatal("不等于1920")
		}
	}
	r.POST(group, "/", fn)
	ts := httptest.NewServer(r.router)
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
	if !strings.EqualFold(string(data), "token need") {
		t.Fatal("没拦截")
	}
	//带token
	id := 1920
	token, err := jwt.CreateToken(id)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", token)
	_, err = (&http.Client{}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
}
