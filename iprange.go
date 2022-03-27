package main

import (
	"encoding/binary"
	"errors"
	"net"
	"strings"
)

type IPRange struct {
	startIP   uint32
	endIP     uint32
	currentIP uint32
}

func NewIpRange(pattern string) (*IPRange, error) {
	var startIP uint32
	var endIP uint32

	if strings.Contains(pattern, "/") {
		// CIDR format: "1.2.3.0/16"
		_, ipnet, err := net.ParseCIDR(pattern)
		if err != nil {
			return nil, err
		}

		mask := binary.BigEndian.Uint32(ipnet.Mask)
		startIP = binary.BigEndian.Uint32(ipnet.IP)
		endIP = (startIP & mask) | (mask ^ 0xffffffff)
	} else if strings.Contains(pattern, "-") {
		// IP range: "1.2.3.1-1.2.3.10"
		ips := strings.Split(pattern, "-")

		startNetIP := net.ParseIP(ips[0])
		if startNetIP == nil {
			return nil, errors.New("Invalid left bound of ip range passed")
		}

		startIP = binary.BigEndian.Uint32(startNetIP[12:16])

		endNetIP := net.ParseIP(ips[1])
		if endNetIP == nil {
			return nil, errors.New("Invalid right bound of ip range passed")
		}

		endIP = binary.BigEndian.Uint32(endNetIP[12:16])

		if startIP > endIP {
			return nil, errors.New("Left bound greater than right bound")
		}
	} else {
		// Single IP: "1.2.3.4"
		netIP := net.ParseIP(pattern)
		if netIP == nil {
			return nil, errors.New("Invalid left bound of ip range passed")
		}

		startIP = binary.BigEndian.Uint32(netIP[12:16])
		endIP = startIP
	}

	ipRange := IPRange{
		startIP:   startIP,
		endIP:     endIP,
		currentIP: startIP,
	}

	return &ipRange, nil
}

func (r *IPRange) Next() {
	r.currentIP = r.currentIP + 1

}

func (r *IPRange) Valid() bool {
	return r.currentIP <= r.endIP
}

func (r *IPRange) Current() net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, r.currentIP)

	return ip
}

func (r *IPRange) Size() uint32 {
	return r.endIP - r.startIP
}
