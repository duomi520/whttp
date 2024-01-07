package whttp

import (
	"encoding/binary"
	"errors"
	"net"

	"net/http"
	"strings"
	"sync"
)

// IPv4子网中，起始IP（最小地址）为该网段的网络地址,如192.168.0.0，结束IP（最大地址）为该网段的广播地址，如192.168.0.255，不可分配给用户使用。如可分配给用户的有效IP范围是192.168.0.1～192.168.0.254。
func calculateCount(ipNet *net.IPNet) int {
	maskLen, _ := ipNet.Mask.Size()
	if maskLen == 32 {
		return 1
	}
	//2的几次方
	count := 1 << (32 - maskLen)
	return count
}

// IPAdmission IP准入
type IPAdmission struct {
	sync.Map
}

func NewIPAdmission() *IPAdmission {
	return &IPAdmission{}
}

var IPAdmissioneEmpty struct{}

// ParseNode 添加规则
func (f *IPAdmission) ParseNode(line string) error {
	if !strings.Contains(line, "/") {
		parsedIP := net.ParseIP(line)
		if ipv4 := parsedIP.To4(); ipv4 != nil {
			parsedIP = ipv4
		}
		if parsedIP != nil {
			switch len(parsedIP) {
			case net.IPv4len:
				key32 := binary.BigEndian.Uint32(parsedIP)
				f.Store(key32, IPAdmissioneEmpty)
			case net.IPv6len:
				f.Store(line, IPAdmissioneEmpty)
			}
		}
	} else {
		baseIP, ipNet, err := net.ParseCIDR(line)
		if err != nil {
			return err
		}
		if ipv4 := baseIP.To4(); ipv4 != nil {
			baseIP = ipv4
		}
		size := calculateCount(ipNet)
		if baseIP != nil {
			switch len(baseIP) {
			case net.IPv4len:
				key32 := binary.BigEndian.Uint32(ipNet.IP.To4())
				for i := 0; i < size; i++ {
					f.Store(key32+uint32(i), IPAdmissioneEmpty)
				}
			case net.IPv6len:
				//TODO
				return errors.New("不支持IPv6 的 CIDR")
			}
		}
	}
	return nil
}

// RemoveNode 移除规则
func (f *IPAdmission) RemoveNode(line string) error {
	if !strings.Contains(line, "/") {
		parsedIP := net.ParseIP(line)
		if ipv4 := parsedIP.To4(); ipv4 != nil {
			parsedIP = ipv4
		}
		if parsedIP != nil {
			switch len(parsedIP) {
			case net.IPv4len:
				key32 := binary.BigEndian.Uint32(parsedIP)
				f.Delete(key32)
			case net.IPv6len:
				f.Delete(line)
			}
		}
	} else {
		baseIP, ipNet, err := net.ParseCIDR(line)
		if err != nil {
			return err
		}
		if ipv4 := baseIP.To4(); ipv4 != nil {
			baseIP = ipv4
		}
		size := calculateCount(ipNet)
		if baseIP != nil {
			switch len(baseIP) {
			case net.IPv4len:
				key32 := binary.BigEndian.Uint32(ipNet.IP.To4())
				for i := 0; i < size; i++ {
					f.Delete(key32 + uint32(i))
				}
			case net.IPv6len:
				//TODO
				return errors.New("不支持IPv6 的 CIDR")
			}
		}
	}
	return nil
}

// Check 检查某个ip在不在设置的规则里
func (f *IPAdmission) Check(line string) any {
	remoteIP := net.ParseIP(line)
	if ipv4 := remoteIP.To4(); ipv4 != nil {
		remoteIP = ipv4
	}
	if remoteIP != nil {
		switch len(remoteIP) {
		case net.IPv4len:
			key32 := binary.BigEndian.Uint32(remoteIP)
			_, ok := f.Load(key32)
			return ok
		case net.IPv6len:
			_, ok := f.Load(line)
			return ok
		}
	}
	return false
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

// BlacklistMiddleware 黑名单。
func (f *IPAdmission) BlacklistMiddleware() func(*HTTPContext) {
	return func(c *HTTPContext) {
		ip := ClientIP(c.Request)
		if !f.Check(ip).(bool) {
			c.Next()
		} else {
			c.String(http.StatusUnauthorized, "拒绝访问")
		}
	}
}
