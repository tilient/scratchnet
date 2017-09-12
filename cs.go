package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

const port = 1876

func client() {
	BROADCAST_IPv4, ip := outboundBroadcastIP()
	addr := &net.UDPAddr{
		IP:   BROADCAST_IPv4,
		Port: port,
	}
	socket, _ := net.DialUDP("udp4", nil, addr)
	data := []byte(ip.String())
	for {
		socket.Write(data)
		time.Sleep(15 * time.Second)
	}
}

func server() {
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
	go client()
	server()
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
	for i, v := range ip {
		bip[i] = v
		mi := i - offset
		if mi >= 0 {
			bip[i] |= ^m[mi]
		}
	}
	return bip
}
