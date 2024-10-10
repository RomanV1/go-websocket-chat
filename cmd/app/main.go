package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/RomanV1/go-websocket-chat/pkg"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

func main() {
	loadEnv()

	hub := initializeHub()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r, hub)
	})

	startServer()
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func initializeHub() *pkg.Hub {
	h := pkg.NewHub()
	go h.Run()
	return h
}

func startServer() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalf("PORT not set in environment")
	}

	log.Printf("Starting the server on :%s", port)
	if err := http.ListenAndServe("localhost:"+port, nil); err != nil {
		log.Fatalf("Error starting the server: %v", err)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, hub *pkg.Hub) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	client := createClient(conn, hub)
	hub.Register <- client

	go client.SendMessages()
	go client.ReceiveMessages()
}

func createClient(conn *websocket.Conn, hub *pkg.Hub) *pkg.Client {
	userUUID := uuid.New()
	client := pkg.NewClient(conn, *hub, userUUID)
	fmt.Printf("User registered: %s \n", userUUID)
	return client
}
