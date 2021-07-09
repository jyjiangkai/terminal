package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"terminal/pkg/eks"
	"terminal/pkg/terminal"
	"terminal/pkg/terminal/websocket"
)

var (
	addr         = flag.String("addr", "0.0.0.0:90", "http service address")
	exec_command = []string{"/bin/sh", "-c", "TERM=xterm-256color; export TERM; [ -x /bin/bash ] && ([ -x /usr/bin/script ] && /usr/bin/script -q -c \"/bin/bash\" /dev/null || exec /bin/bash) || exec /bin/sh"}
)

func ServeWsTerminal(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	namespace := pathParams["namespace"]
	podName := pathParams["pod"]
	containerName := pathParams["container"]
	log.Printf("exec pod: %s, container: %s, namespace: %s\n", podName, containerName, namespace)

	pty, err := websocket.NewTerminalSession(w, r, nil)
	if err != nil {
		log.Printf("get pty failed: %v\n", err)
		return
	}
	defer func() {
		log.Println("close session.")
		pty.Close()
	}()

	eksclient, err := eks.NewClient(r)
	if err != nil {
		log.Printf("create eks client failed, error: %v", err)
		return
	}

	pod, err := eksclient.Get(podName, namespace)
	if err != nil {
		log.Printf("get kubernetes client failed: %v\n", err)
		return
	}
	log.Printf("Get pod: %v\n", pod)

	ok, err := terminal.ValidatePod(pod, containerName)
	if !ok {
		msg := fmt.Sprintf("Validate pod error! err: %v", err)
		log.Println(msg)
		pty.Write([]byte(msg))
		pty.Done()
		return
	}

	err = eksclient.Exec(exec_command, pty, namespace, podName, containerName)
	if err != nil {
		msg := fmt.Sprintf("Exec to pod error! err: %v", err)
		log.Println(msg)
		pty.Write([]byte(msg))
		pty.Done()
	}
	return
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/ecns/pod_exec/{namespace}/{pod}/{container}/", ServeWsTerminal)
	log.Fatal(http.ListenAndServe(*addr, router))
}
