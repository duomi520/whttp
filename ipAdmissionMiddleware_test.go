package whttp

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCalculateCount(t *testing.T) {
	tests := [][2]any{
		{"10.22.2.1/32", 1},
		{"100.111.111.111/22", 1024},
		{"192.168.255.255/28", 16},
		{"192.168.255.255/24", 256},
		{"172.17.0.100/26", 64},
		//	{"2001:db8::/16", 65534},
	}
	for i := range tests {
		_, ipNet, err := net.ParseCIDR(tests[i][0].(string))
		if err != nil {
			t.Log(err)
		}
		v := calculateCount(ipNet)
		if tests[i][1].(int) != v {
			t.Errorf("%s expected %v got %v", tests[i][0], tests[i][1], v)
		}
	}
}
func TestIPAdmission(t *testing.T) {
	list := NewIPAdmission()
	list.ParseNode("127.0.0.1")
	list.ParseNode("192.0.2.5/24")
	list.ParseNode("198.198.110.0/16")
	list.ParseNode("2001:db8::1")
	err := list.ParseNode("2001:db8::/32")
	if err == nil {
		t.Error("IPv6 的CIDR 未识别")
	}
	tests := [][2]any{
		{"127.0.0.1", true},
		{"192.0.2.0", true},
		{"192.0.2.2", true},
		{"192.0.2.255", true},
		{"192.0.3.0", false},
		{"192.0.3.10", false},
		{"198.198.1.100", true},
		{"198.198.111.5", true},
		{"198.199.5.5", false},
		{"2001:db8::1", true},
		{"2001:db9::1", false},
	}
	for i := range tests {
		v := list.Check(tests[i][0].(string)).(bool)
		if tests[i][1].(bool) != v {
			t.Errorf("%s expected %v got %v", tests[i][0], tests[i][1], v)
		}
	}

}

func TestWhitelistMiddleware(t *testing.T) {
	list := NewIPAdmission()
	//list.ParseNode("127.0.0.1")
	list.ParseNode("192.168.0.1")
	fn := func(c *HTTPContext) {
		c.String(200, "拦截失败")
	}
	r := &WRoute{mux: http.NewServeMux()}
	r.GET("/", list.WhitelistMiddleware(), fn)
	ts := httptest.NewServer(r.mux)
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

func BenchmarkCheck(b *testing.B) {
	list := NewIPAdmission()
	list.ParseNode("127.0.0.1")
	list.ParseNode("192.168.0.1")
	list.ParseNode("10.40.68.0/24")
	list.ParseNode("10.40.69.0/24")
	list.ParseNode("10.40.70.0/24")
	for i := 0; i < b.N; i++ {
		list.Check("10.40.68.55")
	}
}
