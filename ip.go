package whttp

import (
	"net"
	"net/http"
	"strings"
)

// HasLocalIPAddr 检测 IP 地址是否是内网地址
// 通过直接对比ip段范围效率更高，详见：https://github.com/thinkeridea/go-extend/issues/2
func HasLocalIPAddr(ip net.IP) bool {
	if ip == nil {
		return false
	}
	if ip.IsLoopback() {
		return true
	}
	// 处理IPv4地址
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] == 10 || // 10.0.0.0/8
			(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) || // 172.16.0.0/12
			(ip4[0] == 169 && ip4[1] == 254) || // 169.254.0.0/16 (APIPA)
			(ip4[0] == 192 && ip4[1] == 168) // 192.168.0.0/16
	}

	// 处理IPv6地址（增加内网地址检测）
	if ip6 := ip.To16(); ip6 != nil {
		// IPv6唯一本地地址 (fc00::/7)
		if ip6[0] == 0xfc || ip6[0] == 0xfd {
			return true
		}

		// IPv6链路本地地址 (fe80::/10)
		if ip6[0] == 0xfe && (ip6[1]&0xc0) == 0x80 {
			return true
		}
	}
	return false
}

// ClientIP 尽最大努力实现获取客户端 IP 的算法。
// 解析 X-Real-IP 和 X-Forwarded-For 以便于反向代理（nginx 或 haproxy）可以正常工作。
func ClientIP(r *http.Request) string {
	ip := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0])
	if ip != "" {
		return ip
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

// ClientPublicIP 尽最大努力实现获取客户端公网 IP 的算法。
// 解析 X-Real-IP 和 X-Forwarded-For 以便于反向代理（nginx 或 haproxy）可以正常工作。
func ClientPublicIP(r *http.Request) string {
	var ip string
	for _, ip = range strings.Split(r.Header.Get("X-Forwarded-For"), ",") {
		if ip = strings.TrimSpace(ip); ip != "" && !HasLocalIPAddr(net.ParseIP(ip)) {
			return ip
		}
	}

	if ip = strings.TrimSpace(r.Header.Get("X-Real-Ip")); ip != "" && !HasLocalIPAddr(net.ParseIP(ip)) {
		return ip
	}

	if ip = RemoteIP(r); !HasLocalIPAddr(net.ParseIP(ip)) {
		return ip
	}

	return ""
}

// RemoteIP 通过 RemoteAddr 获取 IP 地址。
func RemoteIP(r *http.Request) string {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// https://github.com/thinkeridea/go-extend/blob/master/exnet/ip.go
