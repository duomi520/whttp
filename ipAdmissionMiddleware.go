package whttp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"net/http"
	"strings"
	"sync"
)

// IPv4子网中，起始IP（最小地址）为该网段的网络地址,如192.168.0.0，结束IP（最大地址）为该网段的广播地址，如192.168.0.255，不可分配给用户使用。如可分配给用户的有效IP范围是192.168.0.1～192.168.0.254。
func calculateAvailableCount(ipNet *net.IPNet) int {
	ones, bits := ipNet.Mask.Size()
	// IPv4
	if bits == 32 {
		total := 1 << (bits - ones)
		return total
	}
	// IPv6 没有网络/广播地址概念
	return 1 << (bits - ones)
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
		if ipAddr := net.ParseIP(line); ipAddr != nil {
			IPV4 := ipAddr.To4()
			if IPV4 != nil {
				key32 := binary.BigEndian.Uint32(IPV4)
				f.Store(key32, IPAdmissioneEmpty)
			} else {
				f.Store(line, IPAdmissioneEmpty)
			}
			return nil
		}
		return fmt.Errorf("invalid IP: %s", line)
	}
	// CIDR 处理
	baseIP, ipNet, err := net.ParseCIDR(line)
	if err != nil {
		return err
	}
	// IPv4 CIDR
	ipv4 := baseIP.To4()
	size := calculateAvailableCount(ipNet)
	if ipv4 != nil && len(ipv4) == net.IPv4len {
		key32 := binary.BigEndian.Uint32(ipNet.IP.To4())
		for i := range size {
			f.Store(key32+uint32(i), IPAdmissioneEmpty)
		}
		return nil
	}
	return errors.New("CIDR failed")
}

// RemoveNode 移除规则
func (f *IPAdmission) RemoveNode(line string) error {
	if !strings.Contains(line, "/") {
		if ipAddr := net.ParseIP(line); ipAddr != nil {
			IPV4 := ipAddr.To4()
			if IPV4 != nil {
				key32 := binary.BigEndian.Uint32(IPV4)
				f.Delete(key32)
			} else {
				f.Delete(line)
			}
			return nil
		}
		return fmt.Errorf("invalid IP: %s", line)
	}
	// CIDR 处理
	baseIP, ipNet, err := net.ParseCIDR(line)
	if err != nil {
		return err
	}
	// IPv4 CIDR
	ipv4 := baseIP.To4()
	size := calculateAvailableCount(ipNet)
	if ipv4 != nil && len(ipv4) == net.IPv4len {
		key32 := binary.BigEndian.Uint32(ipNet.IP.To4())
		for i := range size {
			f.Delete(key32 + uint32(i))
		}
		return nil
	}
	return errors.New("CIDR failed")
}

// Check 检查某个ip在不在设置的规则里
func (f *IPAdmission) Check(line string) bool {
	remoteIP := net.ParseIP(line)
	if remoteIP == nil {
		return false
	}
	// IPv4 检查
	if ipv4 := remoteIP.To4(); ipv4 != nil {
		key32 := binary.BigEndian.Uint32(ipv4)
		_, ok := f.Load(key32)
		return ok
	}
	// IPv6 检查
	_, ok := f.Load(line)
	return ok
}

// WhitelistMiddleware 白名单。
func (f *IPAdmission) WhitelistMiddleware() func(*HTTPContext) {
	return func(c *HTTPContext) {
		ip := ClientIP(c.Request)
		if f.Check(ip) {
			c.Next()
		} else {
			c.Warn("IP blocked", "ip", ip)
			c.String(http.StatusForbidden, "No access")
		}
	}
}

// BlacklistMiddleware 黑名单。
func (f *IPAdmission) BlacklistMiddleware() func(*HTTPContext) {
	return func(c *HTTPContext) {
		ip := ClientIP(c.Request)
		if !f.Check(ip) {
			c.Next()
		} else {
			c.Warn("IP blocked", "ip", ip)
			c.String(http.StatusForbidden, "No access")
		}
	}
}
