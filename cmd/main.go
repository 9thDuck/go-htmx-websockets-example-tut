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

func main() {
	fmt.Print("Starting system monitor..\n\n")

	go func() {
		for {
			systemSection, err := hardware.GetSystemSection()
			utils.ThrowOnError("systemSection", err)

			diskSection, err := hardware.GetDiskSection()
			utils.ThrowOnError("diskSection", err)

			cpuSection, err := hardware.GetCpuSection()
			utils.ThrowOnError("cpuSection", err)

			fmt.Println(systemSection)
			fmt.Println(diskSection)
			fmt.Println(cpuSection)
			fmt.Print("\n\n")

			time.Sleep(time.Second)
		}
	}()

	srv := NewServer()
	PORT := "8080"
	err := http.ListenAndServe(fmt.Sprintf(":%s", PORT), &srv.mux)
	utils.ThrowOnError("server listening", err)
	fmt.Println("Server listening at")
}
