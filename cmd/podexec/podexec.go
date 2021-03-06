package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	log "k8s.io/klog/v2"
	"net/http"
	"terminal/utils"

	"terminal/pkg/eks"
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
	log.Infof("exec pod: %s, container: %s, namespace: %s\n", podName, containerName, namespace)

	pty, err := websocket.NewTerminalSession(w, r, nil)
	if err != nil {
		log.Errorf("get pty failed: %v\n", err)
		return
	}
	defer func() {
		log.Error("close session.")
		pty.Close()
	}()

	eksclient, err := eks.NewClient(r)
	if err != nil {
		log.Errorf("create eks client failed, error: %v", err)
		return
	}

	pod, err := eksclient.Get(podName, namespace)
	if err != nil {
		log.Errorf("get kubernetes client failed: %v\n", err)
		return
	}
	log.Infof("Get pod: %v\n", pod)

	ok, err := utils.ValidatePod(pod, containerName)
	if !ok {
		msg := fmt.Sprintf("Validate pod error! err: %v", err)
		log.Errorf(msg)
		pty.Write([]byte(msg))
		pty.Done()
		return
	}
	err = eksclient.Exec(exec_command, pty, namespace, podName, containerName)
	if err != nil {
		msg := fmt.Sprintf("Exec to pod error! err: %v", err)
		log.Errorf(msg)
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
