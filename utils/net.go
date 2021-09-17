package utils

import (
	"errors"
	"net"
)

func GetLocalDefaultIp() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iterm := range interfaces {
		if iterm.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iterm.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}

		address, err := iterm.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range address {
			ip := GetIpFromAddr(addr)
			if ip == nil {
				continue
			}
			return ip.String(), nil
		}
	}

	return "", errors.New("找不到本机默认ip")
}

func GetIpFromAddr(addr net.Addr) net.IP {
	var ip net.IP

	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}

	if ip == nil || ip.IsLoopback() {
		return nil
	}

	ip = ip.To4()
	if ip == nil {
		return nil // not an ipv4 address
	}

	return ip
}
