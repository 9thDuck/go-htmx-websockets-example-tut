package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/sigrdrifa/go-htmx-websockets-example/internal/hardware"
	"github.com/sigrdrifa/go-htmx-websockets-example/internal/utils"
)

type subscriber struct {
	msgs chan []byte
}
type server struct {
	subscriberMessageBuffer int
	mux                     http.ServeMux
	subscribersMutex        sync.Mutex
	subscribers             map[*subscriber]struct{}
}

func (s *server) subscribeHandler(writer http.ResponseWriter, req *http.Request) {
	err := s.subscribe(req.Context(), writer, req)
	utils.ThrowOnError("susbcribeHandler", err)
}

func (s *server) subscribe(ctx context.Context, writer http.ResponseWriter, req *http.Request) error {
	var c *websocket.Conn
	subscriber := &subscriber{
		msgs: make(chan []byte, s.subscriberMessageBuffer),
	}
	s.addSubscriber(subscriber)

	c, err := websocket.Accept(writer, req, nil)
	if err != nil {
		return err
	}

	defer c.CloseNow()

	ctx = c.CloseRead(ctx)

	for {
		select {
		case msg := <-subscriber.msgs:
			ctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			err := c.Write(ctx, websocket.MessageText, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *server) addSubscriber(sub *subscriber) {
	s.subscribersMutex.Lock()
	s.subscribers[sub] = struct{}{}
	s.subscribersMutex.Unlock()
	fmt.Println("Added subscriber", *sub)
}

func NewServer() *server {
	s := &server{
		subscriberMessageBuffer: 10,
		subscribers:             make(map[*subscriber]struct{}),
	}
	s.mux.Handle("/", http.FileServer(http.Dir("./htmx")))
	s.mux.HandleFunc("/ws", s.subscribeHandler)

	return s
}

func (s *server) broadcast(msg []byte) {
	s.subscribersMutex.Lock()
	for sub := range s.subscribers {
		sub.msgs <- msg
	}
	s.subscribersMutex.Unlock()
}

func main() {
	fmt.Print("Starting system monitor..\n\n")

	srv := NewServer()

	go func(s *server) {
		for {
			systemSection, err := hardware.GetSystemSection()
			utils.ThrowOnError("systemSection", err)

			diskSection, err := hardware.GetDiskSection()
			utils.ThrowOnError("diskSection", err)

			cpuSection, err := hardware.GetCpuSection()
			utils.ThrowOnError("cpuSection", err)

			timestamp := time.Now().Format("2006-01-02 15:04:05")
			html := `
			<div hx-swap-oob="innerHTML:#update-timestamp">` + timestamp + `</div>
			<div hx-swap-oob="innerHTML:#system-data">` + systemSection + `</div>
			<div hx-swap-oob="innerHTML:#disk-data">` + diskSection + `</div>
			<div hx-swap-oob="innerHTML:#cpu-data">` + cpuSection + `</div>
			`
			s.broadcast([]byte(html))

			time.Sleep(time.Second)
		}
	}(srv)

	PORT := "3001"
	err := http.ListenAndServe(fmt.Sprintf(":%s", PORT), &srv.mux)
	utils.ThrowOnError("server listening", err)
	fmt.Println("Server listening at")
}
