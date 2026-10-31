package daemon

import (
	"net"
	"time"

	"github.com/OoTMM/multiplayer/client/internal/debuglog"
)

/* How long Close waits for queued messages before abandoning the connection. */
const closeDrainTimeout = time.Second

type DaemonConn struct {
	conn   net.Conn
	chSend chan *Msg
	done   chan struct{}
	log    *debuglog.Logger
}

func (conn *DaemonConn) Send(data *Msg) error {
	conn.chSend <- data
	return nil
}

func (d *DaemonConn) sendLoop() {
	defer close(d.done)

	for {
		msg, ok := <-d.chSend
		if !ok {
			return
		}

		/*
			The daemon re-serializes and broadcasts this payload to its websocket
			subscribers; see the debuglog package for the exact guarantees.
		*/
		data, err := sendMsg(d.conn, msg)
		if err != nil {
			d.log.Logf("failed to send %s to daemon: %v", msg.Type, err)
			if data != nil {
				d.log.Log(string(msg.Type), data)
			}
			continue
		}
		d.log.Log(string(msg.Type), data)
	}
}

func createDaemonConn(conn net.Conn, log *debuglog.Logger) *DaemonConn {
	c := &DaemonConn{
		conn:   conn,
		chSend: make(chan *Msg, 128),
		done:   make(chan struct{}),
		log:    log,
	}
	go c.sendLoop()
	return c
}

func (d *DaemonConn) Close() error {
	close(d.chSend)

	/* Let the send loop drain queued messages while the socket is still open. */
	select {
	case <-d.done:
	case <-time.After(closeDrainTimeout):
		d.conn.Close()
		<-d.done
	}

	return d.conn.Close()
}
