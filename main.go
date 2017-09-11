package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
)

func open(url string) error {
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

//------------------------------------------------------------------

var templates = template.Must(template.ParseFiles("app.html"))

func appHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "app.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

//------------------------------------------------------------------

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

//------------------------------------------------------------------

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*default* >>", r.RequestURI)
}

func main() {
	http.HandleFunc("/app", appHandler)
	http.HandleFunc("/poll", pollHandler)
	http.HandleFunc("/reset_all", resetHandler)
	http.HandleFunc("/sendMsg/", sendMsgHandler)
	http.HandleFunc("/waitMsg/", waitMsgHandler)
	http.HandleFunc("/", defaultHandler)

	//go open("http://localhost:56765/view/test")
	log.Fatalln("ListenAndServe:", http.ListenAndServe(":56765", nil))
}
