package pkg

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"strings"
)

type Client struct {
	Hub  Hub
	Conn *websocket.Conn
	Send chan []byte
	ID   uuid.UUID
}

func NewClient(conn *websocket.Conn, hub Hub, id uuid.UUID) *Client {
	return &Client{
		Hub:  hub,
		Conn: conn,
		Send: make(chan []byte),
		ID:   id,
	}
}

func (c *Client) SendMessages() {
	defer func() {
		c.Hub.Unregister <- c
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				log.Printf("error receiving message")
				c.Hub.Unregister <- c
				return
			}

			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("error sending message: %v", err)
				c.Hub.Unregister <- c
				return
			}
		}
	}
}

func (c *Client) ReceiveMessages() {
	defer func() {
		c.Hub.Unregister <- c
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("error receiving message: %v", err)
			return
		}

		if err := c.handleMessage(message); err != nil {
			log.Printf("error handling message: %v", err)
		}
	}
}

func (c *Client) handleMessage(message []byte) error {
	parts := strings.SplitN(string(message), ":", 2)

	switch len(parts) {
	case 2:
		return c.sendMessageToUser(parts[0], parts[1])
	case 1:
		c.Hub.Broadcast <- message
		return nil
	default:
		return fmt.Errorf("invalid message format")
	}
}

func (c *Client) sendMessageToUser(userIDStr, msg string) error {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid UUID format")
	}

	recipient, ok := c.Hub.Clients[userID]
	if !ok {
		return fmt.Errorf("user not found")
	}

	recipient.Send <- []byte(msg)
	return nil
}
