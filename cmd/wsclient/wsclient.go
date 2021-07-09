// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"encoding/json"
	"terminal/pkg/terminal"
)

var addr = flag.String("addr", "127.0.0.1:90", "http service address")

//var addr = flag.String("addr", "10.222.114.241:90", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws/eks/eks-dashboard-api-cfz-debug-57d6b499d7-nxkbr/eks-dashboard-api/webshell"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("[client] recv: %s", message)
		}
	}()

	time.Sleep(5 * time.Second)

	var in string
	msg, err := json.Marshal(terminal.TerminalMessage{
		Operation: "bind",
	})
	log.Printf("[client] send: %v\n", msg)
	err = c.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		log.Println("write:", err)
		return
	}

	msg, err = json.Marshal(terminal.TerminalMessage{
		Operation: "resize",
		Cols:      120,
		Rows:      21,
	})
	log.Printf("[client] send: %v\n", msg)
	err = c.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		log.Println("write:", err)
		return
	}

	in = "whoami"
	msg, err = json.Marshal(terminal.TerminalMessage{
		Operation: "stdin",
		Data:      string(in),
	})
	log.Printf("[client] send: %v\n", msg)
	err = c.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		log.Println("write:", err)
		return
	}

	in = "\r"
	msg, err = json.Marshal(terminal.TerminalMessage{
		Operation: "stdin",
		Data:      string(in),
	})
	log.Printf("[client] send: %v\n", msg)
	err = c.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		log.Println("write:", err)
		return
	}

	in = "pwd\r"
	msg, err = json.Marshal(terminal.TerminalMessage{
		Operation: "stdin",
		Data:      string(in),
	})
	log.Printf("[client] send: %v\n", msg)
	err = c.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		log.Println("write:", err)
		return
	}

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			log.Printf("[client] nothing send")
			//var in string
			//in = "date"
			//msg, err := json.Marshal(terminal.TerminalMessage{
			//      Operation: "stdin",
			//      Data:      string(in),
			//})
			//log.Printf("[client] send: %v\n", msg)
			//err = c.WriteMessage(websocket.TextMessage, []byte(msg))
			//if err != nil {
			//      log.Println("write:", err)
			//      return
			//}
			//
			//in = "\r"
			//msg, err = json.Marshal(terminal.TerminalMessage{
			//      Operation: "stdin",
			//      Data:      string(in),
			//})
			//log.Printf("[client] send: %v\n", msg)
			//err = c.WriteMessage(websocket.TextMessage, []byte(msg))
			//if err != nil {
			//      log.Println("write:", err)
			//      return
			//}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
