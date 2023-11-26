package whttp

import (
	"github.com/duomi520/utils"
	"net"
	"net/http"
	"strings"
)

// IPAdmission IP准入，保存规则nodes
type IPAdmission struct {
	cache *utils.IdempotentCache
	nodes []net.IPNet
}

func NewIPAdmission(power uint64) *IPAdmission {
	a := &IPAdmission{
		nodes: make([]net.IPNet, 0),
	}
	a.cache = utils.NewIdempotentCache(power, 0x0102030405060708, a.check)
	return a

}

// ParseNode 添加规则
func (f *IPAdmission) ParseNode(line string) {
	if !strings.Contains(line, "/") {
		parsedIP := net.ParseIP(line)
		if ipv4 := parsedIP.To4(); ipv4 != nil {
			parsedIP = ipv4
		}
		if parsedIP != nil {
			switch len(parsedIP) {
			case net.IPv4len:
				line += "/32"
			case net.IPv6len:
				line += "/128"
			}
		}
	}
	_, cidrNet, err := net.ParseCIDR(line)
	if err == nil {
		f.nodes = append(f.nodes, *cidrNet)
	}
}

func (f *IPAdmission) check(ip []byte) any {
	remoteIP := net.ParseIP(utils.BytesToString(ip))
	for i := range f.nodes {
		if f.nodes[i].Contains(remoteIP) {
			return true
		}
	}
	return false
}

// Check 检查某个ip在不在设置的规则里
func (f *IPAdmission) Check(ip string) bool {
	return f.cache.Get(utils.StringToBytes(ip)).(bool)
}

// WhitelistMiddleware 白名单
func (f *IPAdmission) WhitelistMiddleware() func(*HTTPContext) {
	return func(c *HTTPContext) {
		ip := ClientIP(c.Request)
		if f.Check(ip) {
			c.Next()
		} else {
			c.String(http.StatusUnauthorized, "拒绝访问")
		}
	}
}

// BlacklistMiddleware 黑名单
func (f *IPAdmission) BlacklistMiddleware() func(*HTTPContext) {
	return func(c *HTTPContext) {
		ip := ClientIP(c.Request)
		if !f.Check(ip) {
			c.Next()
		} else {
			c.String(http.StatusUnauthorized, "拒绝访问")
		}
	}
}
