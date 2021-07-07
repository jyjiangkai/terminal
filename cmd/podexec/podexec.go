package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"terminal/pkg/eks"
	_ "terminal/initialize"
	"terminal/pkg/session"
	"terminal/pkg/terminal"
	"terminal/pkg/terminal/websocket"
	"terminal/utils"
)


var (
	addr = flag.String("addr", "0.0.0.0:90", "http service address")
	//cmd  = []string{"/bin/sh"}
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

	// Get project id form cookies, use ems api
	sessions, err := session.EmsSessionAuth(r)
	if err != nil {
		log.Printf("get session failed: %v\n", err)
		return
	}

	// Get cluster info from cluster name and project id
	cluster, err := utils.GetClusterInfo("redisnotdelete", sessions.ProjectID)
	if err != nil {
		log.Printf("get cluster info: cluster %s, projectID %s, error: %v", "redisnotdelete", sessions.ProjectID, err)
		return
	}

	// Get token
	eksToken, err := utils.GetToken(r, cluster, sessions.ProjectID)
	if err != nil {
		log.Printf("get or validate cluster token: cluster %s %s, projectID %s, error: %v", *cluster.Name, *cluster.APIServerAddress, sessions.ProjectID, err)
		return
	}

	eksclient, err := eks.NewEKSClient(*cluster.APIServerAddress, eksToken)
	if err != nil {
		log.Printf("create eks client: api server: %s, token: %s, error: %v", *cluster.APIServerAddress, eksToken, err)
		return
	}

	pod, err := eksclient.Get(podName, namespace)
	if err != nil {
		log.Printf("get kubernetes client failed: %v\n", err)
		return
	}


	//client, err := kube.GetClient()
	//if err != nil {
	//	log.Printf("get kubernetes client failed: %v\n", err)
	//	return
	//}
	//pod, err := client.PodBox.Get(podName, namespace)
	//if err != nil {
	//	log.Printf("get kubernetes client failed: %v\n", err)
	//	return
	//}
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