package main

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type client struct {
	userID  int
	socket  *websocket.Conn
	receive chan Message
	room    *room
}

func (c *client) read() {
	defer c.socket.Close()
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}
		message := Message{UserID: c.userID}
		err = json.Unmarshal(msg, &message)
		if err != nil {
			log.Println(err)
			continue
		}
		c.room.forward <- message
	}
}

func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.receive {
		err := c.socket.WriteMessage(websocket.TextMessage, msg.Bytes())
		if err != nil {
			return
		}
	}
}
