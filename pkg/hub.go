package pkg

import "github.com/google/uuid"

type Hub struct {
	Clients    map[uuid.UUID]*Client
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[uuid.UUID]*Client),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case newClient := <-h.Register:
			h.Clients[newClient.ID] = newClient
		case newClient := <-h.Unregister:
			if _, ok := h.Clients[newClient.ID]; ok {
				delete(h.Clients, newClient.ID)
				close(newClient.Send)
			}
		case message := <-h.Broadcast:
			for id, client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, id)
				}
			}
		}
	}
}
