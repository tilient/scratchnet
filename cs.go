package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

const port = 56865

func sender() {
	bip, ip := outboundBroadcastIP()
	addr := &net.UDPAddr{
		IP:   bip,
		Port: port,
	}
	socket, _ := net.DialUDP("udp4", nil, addr)
	data := []byte(ip.String())
	for {
		socket.Write(data)
		time.Sleep(15 * time.Second)
	}
}

func listener() {
	socket, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: port,
	})
	if err != nil {
		log.Fatal(err)
	}
	for {
		data := make([]byte, 16)
		n, _, _ := socket.ReadFromUDP(data)
		fmt.Println(" str(data):", string(data[:n]))
	}
}

func main() {
	bip, ip := outboundBroadcastIP()
	fmt.Println("broadcast IP =", bip, " IP =", ip)
	go sender()
	listener()
}

func outboundBroadcastIP() (net.IP, net.IP) {
	outboundIP := GetOutboundIP()
	interfaces, _ := net.Interfaces()
	for _, intf := range interfaces {
		addrs, _ := intf.Addrs()
		for _, addr := range addrs {
			ipv4Addr, _, _ := net.ParseCIDR(addr.String())
			if ipv4Addr.String() == outboundIP.String() {
				return broadcastIP(addr.String()), outboundIP
			}
		}
	}
	return nil, outboundIP
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func broadcastIP(s string) net.IP {
	i := strings.Index(s, "/")
	addr, mask := s[:i], s[i+1:]
	ip := net.ParseIP(addr)
	n, _ := strconv.Atoi(mask)
	m := net.CIDRMask(n, 8*net.IPv4len)
	bip := make(net.IP, len(ip))
	offset := len(ip) - len(m)
	for ix, v := range ip {
		bip[ix] = v
		if ix >= offset {
			bip[ix] |= ^m[ix-offset]
		}
	}
	return bip
}
