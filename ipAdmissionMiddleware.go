package whttp

import (
	"github.com/duomi520/utils"
	"net"
	"net/http"
	"strings"
	"sync"
)

// IPAdmission IP准入，保存规则nodes
type IPAdmission struct {
	cache *utils.IdempotentCache[string]
	nodes []net.IPNet
	sync.RWMutex
}

func NewIPAdmission(power uint64) *IPAdmission {
	a := &IPAdmission{
		cache: &utils.IdempotentCache[string]{},
		nodes: make([]net.IPNet, 0),
	}
	a.cache.Init(power, 0x0102030405060708, a.Check)
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
		f.Lock()
		f.nodes = append(f.nodes, *cidrNet)
		f.Unlock()
	}
}

// Check 检查某个ip在不在设置的规则里
func (f *IPAdmission) Check(ip string) any {
	remoteIP := net.ParseIP(ip)
	f.RLock()
	defer f.RUnlock()
	for i := range f.nodes {
		if f.nodes[i].Contains(remoteIP) {
			return true
		}
	}
	return false
}

func (f *IPAdmission) CheckByCache(ip string) bool {
	return f.cache.Get(ip).(bool)
}

// WhitelistMiddleware 白名单。
func (f *IPAdmission) WhitelistMiddleware() func(*HTTPContext) {
	return func(c *HTTPContext) {
		ip := ClientIP(c.Request)
		if f.Check(ip).(bool) {
			c.Next()
		} else {
			c.String(http.StatusUnauthorized, "拒绝访问")
		}
	}
}

// WhitelistByCacheMiddleware 缓存白名单,规则启用后不能更改。
func (f *IPAdmission) WhitelistByCacheMiddleware() func(*HTTPContext) {
	return func(c *HTTPContext) {
		ip := ClientIP(c.Request)
		if f.CheckByCache(ip) {
			c.Next()
		} else {
			c.String(http.StatusUnauthorized, "拒绝访问")
		}
	}
}
