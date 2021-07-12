// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	log "k8s.io/klog/v2"
	"net/http"
	_ "net/http/pprof"
	"terminal/utils"

	"github.com/gorilla/mux"
	"terminal/pkg/kube"
	wsterminal "terminal/pkg/terminal/websocket"
)

var (
	addr = flag.String("addr", ":8090", "http service address")
	cmd  = []string{"/bin/sh"}
)

func serveTerminal(w http.ResponseWriter, r *http.Request) {
	// auth
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "./frontend/terminal.html")
}

func serveLogs(w http.ResponseWriter, r *http.Request) {
	// auth
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "./frontend/logs.html")
}

func serveWsTerminal(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	namespace := pathParams["namespace"]
	podName := pathParams["pod"]
	containerName := pathParams["container"]
	log.Infof("exec pod: %s, container: %s, namespace: %s\n", podName, containerName, namespace)

	pty, err := wsterminal.NewTerminalSession(w, r, nil)
	if err != nil {
		log.Errorf("get pty failed: %v\n", err)
		return
	}
	defer func() {
		log.Infof("close session.")
		pty.Close()
	}()

	client, err := kube.NewClient()
	if err != nil {
		log.Errorf("get kubernetes client failed: %v\n", err)
		return
	}
	pod, err := client.PodBox.Get(podName, namespace)
	if err != nil {
		log.Errorf("get kubernetes client failed: %v\n", err)
		return
	}
	ok, err := utils.ValidatePod(pod, containerName)
	if !ok {
		msg := fmt.Sprintf("Validate pod error! err: %v", err)
		log.Errorf(msg)
		pty.Write([]byte(msg))
		pty.Done()
		return
	}
	err = client.PodBox.Exec(cmd, pty, namespace, podName, containerName)
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
	router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	// TODO
	// temporarily use relative path, run by `go run cmd/webshell/webshell_main.go` in project root path.
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./frontend/"))))
	// enter webshell by url like: http://127.0.0.1:8090/terminal?namespace=default&pod=nginx-65f9798fbf-jdrgl&container=nginx
	router.HandleFunc("/terminal", serveTerminal)
	router.HandleFunc("/ws/{namespace}/{pod}/{container}/webshell", serveWsTerminal)
	router.HandleFunc("/logs", serveLogs)
	log.Fatal(http.ListenAndServe(*addr, router))
}
