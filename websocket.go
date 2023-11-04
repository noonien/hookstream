package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

const (
	pingPeriod      = 50 * time.Second
	inactiveTimeout = 2 * time.Minute
	writeTimeout    = 5 * time.Second
)

var upgrader = websocket.Upgrader{}

func handleSocket(w http.ResponseWriter, r *http.Request) {
	topic := chi.URLParam(r, "topic")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		writeError(w, r, http.StatusUpgradeRequired, err)
		return
	}

	wss := &wsSub{sendC: make(chan []byte, 20)}
	sub(topic, wss)

	go wss.handle(conn, topic)
}

type wsSub struct {
	sendC chan []byte
}

func (wss *wsSub) Send(data []byte) { wss.sendC <- data }

// handle manages the lifecycle of a websocket connection and ensures the subscriber is unsubscribed on closure.
func (wss *wsSub) handle(conn *websocket.Conn, topic string) {
	defer conn.Close()
	defer unsub(topic, wss)

	// this channel is used to signal read operations success
	// receiving a nil error resets the inactivity timer
	// receiving an error closes the connection
	readerr := make(chan error, 10)

	// start reading from the connection, this is required to read pongs
	// and to detect connection closure
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			readerr <- err
			if err != nil {
				return
			}
		}
	}()

	// receiving a ping resets the inactivity timeout
	conn.SetPongHandler(func(string) error {
		readerr <- nil
		return nil
	})

	pingt := time.NewTicker(pingPeriod)
	timeout := time.NewTicker(inactiveTimeout)

	for {
		select {
		// send published data
		case data := <-wss.sendC:
			if err := wsWrite(conn, websocket.TextMessage, data); err != nil {
				return
			}

		// check for read error
		// reset timers on nil, error
		case rerr := <-readerr:
			if rerr != nil {
				return
			}
			pingt.Reset(pingPeriod)
			timeout.Reset(inactiveTimeout)

		// send a ping, reset inactivity timer
		case <-pingt.C:
			if err := wsWrite(conn, websocket.PingMessage, nil); err != nil {
				return
			}

		// abort on timeout
		case <-timeout.C:
			return
		}
	}
}

// wsWrite writes a message to the websocket connection with a deadline
func wsWrite(conn *websocket.Conn, mtype int, data []byte) error {
	if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		return err
	}
	return conn.WriteMessage(mtype, data)
}
