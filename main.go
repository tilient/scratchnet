package main

import (
	"bufio"
	"context"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/GeertJohan/go.rice"
	"golang.org/x/net/websocket"
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

var (
	peers      map[string]int = make(map[string]int)
	peersMutex sync.Mutex
)

func main() {
	addr, _ := net.ResolveUDPAddr("udp4", "224.0.0.1:56865")
	conn, _ := net.ListenMulticastUDP("udp4", nil, addr)
	go func() {
		b := make([]byte, 2048)
		for {
			n, raddr, _ := conn.ReadFromUDP(b)
			host, _, _ := net.SplitHostPort(raddr.String())
			fmt.Println("host: ", host, " >> ", string(b[:n]))
			peersMutex.Lock()
			peers[host] = 60
			peersMutex.Unlock()
		}
	}()

	ad, _ := net.ResolveUDPAddr("udp4", "224.0.0.1:56865")
	c, _ := net.DialUDP("udp4", nil, ad)
	msg := []byte("scratchnet")
	go func() {
		for {
			c.Write(msg)
			time.Sleep(5 * time.Second)
		}
	}()
	for {
		fmt.Println("peers:")
		peersMutex.Lock()
		peers["127.0.0.1"] = 60
		for k, v := range peers {
			fmt.Println(" ", k, " - ", v)
			if v <= 0 {
				delete(peers, k)
			} else {
				peers[k] = v - 15
			}
		}
		peersMutex.Unlock()
		time.Sleep(15 * time.Second)
	}
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

func openWebview() {
	title := "Scratch Net"
	url := "http://localhost:56765/app"
	w, h := 640, 350
	if runtime.GOOS == "windows" {
		w, h = 648, 386
	}
	titleStr := C.CString(title)
	defer C.free(unsafe.Pointer(titleStr))
	urlStr := C.CString(url)
	defer C.free(unsafe.Pointer(urlStr))
	resize := C.int(1)
	r := C.webview(titleStr, urlStr,
		C.int(w), C.int(h), resize)
	if r != 0 {
		log.Fatal("--5--", "failed to create webview")
	}
}

//---------------------------------------------------------

var appTmpl *template.Template = nil

func appHandler(w http.ResponseWriter, r *http.Request) {
	if appTmpl == nil {
		box, err := rice.FindBox("www")
		if err != nil {
			log.Fatal("--6--", err)
		}
		appStr, err := box.String("app.html")
		if err != nil {
			log.Fatal("--7--", err)
		}
		appTmpl, err = template.New("app").Parse(appStr)
		if err != nil {
			log.Fatal("--8--", err)
		}
	}
	err := appTmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
	}
}

//---------------------------------------------------------

var serv *http.Server

var connections map[*websocket.Conn]bool = make(map[*websocket.Conn]bool)

var msgs map[string]string = make(map[string]string)
var wmsgs map[string]string = make(map[string]string)

//---------------------------------------------------------

func wsServer(ws *websocket.Conn) {
	connections[ws] = true
	scanner := bufio.NewScanner(ws)
	for scanner.Scan() {
		msg := scanner.Text()
		switch msg {
		case "exit":
			exit()
		case "openurl":
			wsOpenUrl(scanner)
		case "openapp":
			go openWebview()
		default:
			fmt.Println("ERROR: unknown:", msg)
			ws.Write([]byte(
				"alert('ERROR: unknown: " + msg + "')"))
		}
	}
	delete(connections, ws)
}

func wsOpenUrl(s *bufio.Scanner) {
	if s.Scan() {
		url := s.Text()
		openUrl("http://localhost:56765" + url)
	}
}

func openAppHandler(w http.ResponseWriter, r *http.Request) {
	if len(connections) < 1 {
		go openWebview()
	}
}

//---------------------------------------------------------

func pollHandler(w http.ResponseWriter, r *http.Request) {
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
	msgs = make(map[string]string)
	wmsgs = make(map[string]string)
	fmt.Fprintln(w)
}

func sendMsgBasicHandler(w http.ResponseWriter,
	r *http.Request) {
	parts := strings.Split(r.RequestURI, "/")
	msg := parts[1]
	to := parts[2]
	msgs[to] = msg
	delete(wmsgs, to)
}

func sendMsgHandler(w http.ResponseWriter,
	r *http.Request) {
	parts := strings.Split(r.RequestURI, "/")
	msg := parts[2]
	to := parts[3]
	peersMutex.Lock()
	for ip, _ := range peers {
		resp, err := http.Get(
			"http://" + ip + ":56765/sendMsgBasic/" + msg + "/" + to)
		if err == nil {
			resp.Body.Close()
		}
	}
	peersMutex.Unlock()
}

func waitMsgHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.RequestURI, "/")
	id := parts[2]
	from := parts[3]
	wmsgs[from] = id
}

//---------------------------------------------------------

func sendAll(str string) {
	for ws, _ := range connections {
		ws.Write([]byte(str))
	}
}

func exit() {
	sendAll("window.close()")
	ctx, _ := context.WithTimeout(
		context.Background(), 1*time.Second)
	serv.Shutdown(ctx)
	os.Exit(0)
}

func mainXX() {
	resp, err := http.Get("http://localhost:56765/openapp")
	if err == nil {
		resp.Body.Close()
		os.Exit(0)
	}

	http.HandleFunc("/app", appHandler)
	http.HandleFunc("/openapp", openAppHandler)
	http.Handle("/ws", websocket.Handler(wsServer))
	http.Handle("/", http.FileServer(
		rice.MustFindBox("www").HTTPBox()))

	http.HandleFunc("/poll", pollHandler)
	http.HandleFunc("/reset_all", resetHandler)
	http.HandleFunc("/sendMsg/", sendMsgHandler)
	http.HandleFunc("/sendMsgBasic/", sendMsgBasicHandler)
	http.HandleFunc("/waitMsg/", waitMsgHandler)

	serv = &http.Server{
		Addr:           ":56765",
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go openWebview()

	log.Fatal("--10--", serv.ListenAndServe())
}

//---------------------------------------------------------
