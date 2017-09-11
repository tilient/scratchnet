package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	fmt.Println(localAddr)
	return localAddr.IP
}

const port = 1876

func client() {
	GetOutboundIP()
	BROADCAST_IPv4 := net.IPv4(192, 168, 0, 255)
	addr := &net.UDPAddr{
		IP:   BROADCAST_IPv4,
		Port: port,
	}
	socket, _ := net.DialUDP("udp4", nil, addr)
	data := []byte("aaaaaaaaaa\n")
	for {
		time.Sleep(5000 * time.Millisecond)
		socket.Write(data)
		fmt.Print(".")
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
		_, remoteAddr, _ := socket.ReadFromUDP(data)
		fmt.Println("remoteAddr:", remoteAddr)
		fmt.Println("      data:", data)
		fmt.Println("    string(data):", string(data))
	}
}

func main() {
	go client()
	server()
}
