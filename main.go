package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"unsafe"
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

var appStr string = `<html>
  <head>
    <title>Scratch Net</title>
		<style type="text/css">
		body {
	    background: url("/imgs/bg01.jpg") no-repeat center fixed;
			background-size: cover;
	  }
		</style>
  </head>
  <body>
    <h1>ScratchNet</h1>
    <div>een eerste testje ...</div>
    <hr />
    <a href="/exit">exit</a>
    <hr />
  </body>
</html>`

var appTmpl *template.Template = nil

func appHandler(w http.ResponseWriter, r *http.Request) {
	if appTmpl == nil {
		appTmpl, _ = template.New("app").Parse(appStr)
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

func exit() {
	ctx, _ := context.WithTimeout(
		context.Background(), 2*time.Millisecond)
	serv.Shutdown(ctx)
	os.Exit(0)
}

func exitHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*exit* >>", r.RequestURI)
	exit()
}

func main() {
	http.HandleFunc("/app", appHandler)
	http.HandleFunc("/poll", pollHandler)
	http.HandleFunc("/reset_all", resetHandler)
	http.HandleFunc("/sendMsg/", sendMsgHandler)
	http.HandleFunc("/waitMsg/", waitMsgHandler)
	http.HandleFunc("/exit", exitHandler)
	http.Handle("/", http.StripPrefix("/",
		http.FileServer(http.Dir("."))))

	serv = &http.Server{
		Addr:           ":56763",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		openWebview("Scratch Net", "http://localhost:56763/app",
			600, 400)
	}()
	log.Fatal(serv.ListenAndServe())
}

//---------------------------------------------------------
