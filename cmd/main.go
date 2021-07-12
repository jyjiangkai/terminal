package main

import (
	"flag"
	"fmt"
	"net/http"

	log "k8s.io/klog/v2"

	"github.com/gorilla/mux"

	"terminal/pkg/eks"
	"terminal/pkg/terminal/websocket"
	"terminal/utils"
)

var (
	addr        = flag.String("addr", "0.0.0.0:90", "http service address")
	execCommand = []string{"/bin/sh", "-c", "TERM=xterm-256color; export TERM; [ -x /bin/bash ] && ([ -x /usr/bin/script ] && /usr/bin/script -q -c \"/bin/bash\" /dev/null || exec /bin/bash) || exec /bin/sh"}
)

func ServeWsTerminal(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	namespace := pathParams["namespace"]
	podName := pathParams["pod"]
	containerName := pathParams["container"]
	log.Infof("Received pod exec request, namespace: %s, pod: %s, container: %s", namespace, podName, containerName)

	pty, err := websocket.NewTerminalSession(w, r, nil)
	if err != nil {
		log.Errorf("New websocket terminal session failed, err: %v", err)
		return
	}
	defer func() {
		log.Info("Close websocket terminal session.")
		pty.Close()
	}()
	log.Info("New websocket terminal session success.")

	eksclient, err := eks.NewClient(r)
	if err != nil {
		log.Errorf("New eks client failed, err: %v", err)
		return
	}
	log.Info("New websocket terminal session success.")

	pod, err := eksclient.Get(podName, namespace)
	if err != nil {
		log.Errorf("Get eks pod failed, pod: %s, err: %v", podName, err)
		return
	}
	log.Infof("Get eks pod success, pod: %v", pod)

	ok, err := utils.ValidatePod(pod, containerName)
	if !ok {
		msg := fmt.Sprintf("Validate pod failed, pod: %v, container: %s, err: %v", pod, containerName, err)
		log.Errorf(msg)
		pty.Write([]byte(msg))
		pty.Done()
		return
	}
	log.Infof("Validate pod success, pod: %v\n", pod)

	err = eksclient.Exec(execCommand, pty, namespace, podName, containerName)
	if err != nil {
		msg := fmt.Sprintf("Exec to eks pod failed, namespace: %s, pod: %s, container: %s, err: %v", namespace, podName, containerName, err)
		log.Errorf(msg)
		pty.Write([]byte(msg))
		pty.Done()
	}
	log.Infof("Exec to eks pod success, namespace: %s, pod: %s, container: %s", namespace, podName, containerName)
	return
}

func main() {
	i := 5
	fmt.Println(i)
	router := mux.NewRouter()
	router.HandleFunc("/api/ecns/pod_exec/{namespace}/{pod}/{container}/", ServeWsTerminal)
	log.Fatal(http.ListenAndServe(*addr, router))
}
