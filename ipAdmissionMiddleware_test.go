package whttp

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestIPAdmission(t *testing.T) {
	list := NewIPAdmission(10)
	list.ParseNode("127.0.0.1")
	list.ParseNode("192.0.2.1/24")
	list.ParseNode("2001:db8::/32")
	tests := [][2]any{
		{"127.0.0.1", true},
		{"192.0.2.10", true},
		{"192.0.3.10", false},
		{"2001:db8::1", true},
		{"2001:db9::1", false},
	}
	for i := range tests {
		data := list.Check(tests[i][0].(string))
		if tests[i][1].(bool) != bool(data) {
			t.Errorf("%s expected %v got %v", tests[i][0], tests[i][1], data)
		}
	}

}

func TestWhitelistMiddleware(t *testing.T) {
	list := NewIPAdmission(10)
	//list.ParseNode("127.0.0.1")
	list.ParseNode("192.168.0.1")
	fn := func(c *HTTPContext) {
		c.String(200, "拦截失败")
	}
	r := &WRoute{router: httprouter.New()}
	r.GET("/", list.WhitelistMiddleware(), fn)
	ts := httptest.NewServer(r.router)
	defer ts.Close()
	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(data, []byte("拦截失败")) {
		t.Error(string(data))
	}
}

func BenchmarkIPFiltering(b *testing.B) {
	list := NewIPAdmission(10)
	list.ParseNode("127.0.0.1")
	list.ParseNode("192.168.0.1")
	list.ParseNode("10.40.68.0/24")
	list.ParseNode("10.40.69.0/24")
	list.ParseNode("10.40.70.0/24")
	for i := 0; i < b.N; i++ {
		list.Check("10.40.68.55")
	}
}
