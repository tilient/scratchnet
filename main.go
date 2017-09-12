package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/GeertJohan/go.rice"
)

/*
#cgo linux CFLAGS: -DWEBVIEW_GTK=1
#cgo linux pkg-config: gtk+-3.0 webkitgtk-3.0

#cgo windows CFLAGS: -DWEBVIEW_WINAPI=1
#cgo windows LDFLAGS: -lole32 -lcomctl32 -loleaut32 -luuid -mwindows

#cgo darwin CFLAGS: -DWEBVIEW_COCOA=1 -x objective-c
#cgo darwin LDFLAGS: -framework Cocoa -framework WebKit

#include <stdlib.h>
#include "webview.h"
*/
import "C"

//---------------------------------------------------------

func openWebview(title, url string, w, h int) error {
	titleStr := C.CString(title)
	defer C.free(unsafe.Pointer(titleStr))
	urlStr := C.CString(url)
	defer C.free(unsafe.Pointer(urlStr))
	resize := C.int(1)
	r := C.webview(titleStr, urlStr,
		C.int(w), C.int(h), resize)
	if r != 0 {
		return errors.New("failed to create webview")
	}
	fmt.Println("* webview opened *")
	exit()
	return nil
}

//---------------------------------------------------------

var appTmpl *template.Template = nil

func appHandler(w http.ResponseWriter, r *http.Request) {
	if appTmpl == nil {
		box, err := rice.FindBox("www")
		if err != nil {
			log.Fatal(err)
		}
		appStr, err := box.String("app.html")
		if err != nil {
			log.Fatal(err)
		}
		appTmpl, err = template.New("app").Parse(appStr)
		if err != nil {
			log.Fatal(err)
		}
	}
	err := appTmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
	}
}

//---------------------------------------------------------

var msgs map[string]string = make(map[string]string)
var wmsgs map[string]string = make(map[string]string)

func pollHandler(w http.ResponseWriter, r *http.Request) {
	//fmt.Print(".")
	for k, v := range msgs {
		fmt.Fprintf(w, "readMsg/%s %s\n", k, v)
	}
	for _, v := range wmsgs {
		fmt.Fprintf(w, "_busy %s\n", v)
	}
	fmt.Fprintln(w)
	msgs = make(map[string]string)
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*reset*", r.RequestURI)
	msgs = make(map[string]string)
	fmt.Fprintln(w)
}

func sendMsgHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*sendMsg*", r.RequestURI)
	parts := strings.Split(r.RequestURI, "/")
	msg := parts[2]
	to := parts[3]
	msgs[to] = msg
	delete(wmsgs, to)
}

func waitMsgHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*waitMsg*", r.RequestURI)
	parts := strings.Split(r.RequestURI, "/")
	id := parts[2]
	from := parts[3]
	wmsgs[from] = id
}

//---------------------------------------------------------

var serv *http.Server

func openUrlHandler(w http.ResponseWriter, r *http.Request) {
	url := strings.TrimPrefix(r.RequestURI, "/openurl")
	url = "http://localhost:56765" + url
	fmt.Println("*openUrl*", r.RequestURI, url)
	openUrl(url)
	http.Redirect(w, r, "/app", http.StatusFound)
}

func exit() {
	ctx, _ := context.WithTimeout(
		context.Background(), 1*time.Second)
	serv.Shutdown(ctx)
	os.Exit(0)
}

func exitHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*exit* >>", r.RequestURI)
	http.Redirect(w, r,
		"http://tilient.github.io/scratchnet/", http.StatusFound)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	exit()
}

func main() {
	http.HandleFunc("/app", appHandler)
	http.HandleFunc("/openurl/", openUrlHandler)
	http.HandleFunc("/poll", pollHandler)
	http.HandleFunc("/reset_all", resetHandler)
	http.HandleFunc("/sendMsg/", sendMsgHandler)
	http.HandleFunc("/waitMsg/", waitMsgHandler)
	http.HandleFunc("/exit", exitHandler)
	//http.Handle("/", http.FileServer(http.Dir("www")))

	http.Handle("/", http.FileServer(
		rice.MustFindBox("www").HTTPBox()))

	serv = &http.Server{
		Addr:           ":56765",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		openWebview("Scratch Net", "http://localhost:56765/app",
			664, 370)
	}()
	log.Fatal(serv.ListenAndServe())
}

//---------------------------------------------------------

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

//---------------------------------------------------------

func openUrl(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

//---------------------------------------------------------
