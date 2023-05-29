package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"

	"github.com/gorilla/websocket"
)

type Message struct {
	Text   string `json:"text"`
	RoomID int    `json:"room_id"`
	UserID int    `json:"user_id"`
}

func (m Message) Bytes() []byte {
	b, _ := json.Marshal(m)
	return b
}

type room struct {

	// clients holds all current clients in this room.
	clients map[*client]bool

	// join is a channel for clients wishing to join the room.
	join chan *client

	// leave is a channel for clients wishing to leave the room.
	leave chan *client

	// forward is a channel that holds incoming messages that should be forwarded to the other clients.
	forward chan Message
}

// newRoom create a new chat room

func newRoom() *room {
	return &room{
		forward: make(chan Message),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

var messages []Message

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
			for _, message := range messages {
				client.receive <- message
			}
			for c := range r.clients {
				c.receive <- Message{Text: "join", UserID: client.userID}
			}

		case client := <-r.leave:
			delete(r.clients, client)
			for c := range r.clients {
				c.receive <- Message{Text: "left the chat", UserID: client.userID}
			}
			close(client.receive)
		case msg := <-r.forward:
			for c := range r.clients {
				c.receive <- msg
			}
			messages = append(messages, msg)
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}
	client := &client{
		socket:  socket,
		receive: make(chan Message, messageBufferSize),
		room:    r,
		userID:  rand.Intn(10000) + 1,
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
