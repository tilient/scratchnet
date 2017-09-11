package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

const port = 1876

func client() {
	BROADCAST_IPv4 := net.IPv4(255, 255, 255, 255)
	addr := &net.UDPAddr{
		IP:   BROADCAST_IPv4,
		Port: port,
	}
	for {
		socket, _ := net.DialUDP("udp4", nil, addr)
		time.Sleep(5000 * time.Millisecond)
		data := []byte("blah")
		socket.WriteToUDP(data, addr)
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
	}
}

func main() {
	go client()
	server()
}
