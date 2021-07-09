package terminal

import (
	"io"
	"time"

	"k8s.io/client-go/tools/remotecommand"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 8192

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Time to wait before force close on connection.
	closeGracePeriod = 10 * time.Second

	// EndOfTransmission end
	EndOfTransmission = "\u0004"
)

type TerminalMessage struct {
	Operation string `json:"Op"`
	Data      string `json:"Data"`
	Rows      uint16 `json:"Rows"`
	Cols      uint16 `json:"Cols"`
}

// PtyHandler is what remotecommand expects from a pty
type PtyHandler interface {
	remotecommand.TerminalSizeQueue
	Done()
	Tty() bool
	Stdin() io.Reader
	Stdout() io.Writer
	Stderr() io.Writer
}
