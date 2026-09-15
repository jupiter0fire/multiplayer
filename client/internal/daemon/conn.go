package daemon

import (
	"net"

	"github.com/OoTMM/multiplayer/client/internal/debuglog"
)

type DaemonConn struct {
	conn   net.Conn
	chSend chan *Msg
	log    *debuglog.Logger
}

func (conn *DaemonConn) Send(data *Msg) error {
	conn.chSend <- data
	return nil
}

func (d *DaemonConn) sendLoop() {
	for {
		msg, ok := <-d.chSend
		if !ok {
			return
		}

		/* The daemon broadcasts exactly this payload to its websocket subscribers */
		data, err := sendMsg(d.conn, msg)
		if err != nil {
			d.log.Logf("failed to send %s to daemon: %v", msg.Type, err)
			continue
		}
		d.log.Log(string(msg.Type), data)
	}
}

func createDaemonConn(conn net.Conn, log *debuglog.Logger) *DaemonConn {
	c := &DaemonConn{
		conn:   conn,
		chSend: make(chan *Msg, 128),
		log:    log,
	}
	go c.sendLoop()
	return c
}

func (d *DaemonConn) Close() error {
	close(d.chSend)
	return d.conn.Close()
}
