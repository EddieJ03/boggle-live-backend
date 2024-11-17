package main

import (
	"fmt"

	"math/rand"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

var (
	clientRooms = make(map[string]*Room)
	clientRoomsLock sync.RWMutex

	randomRooms = make([]*Room, 0, 10)
	randomRoomsLock sync.Mutex
)

const NUM = 4

var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow all origins
    },
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	wsClient := &WSClient{
		Conn: conn, 
		RoomName: "", 
		UniqueNumber: rand.Int(), 
		Number: -1, 
	}

	wsClient.HandleClient()
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleConnections)
	handler := cors.Default().Handler(mux)

	server := &http.Server{
		Addr:    ":5000",
		Handler: handler,
	}

	fmt.Println("Server is running on port 5000!")

	server.ListenAndServe()
}

