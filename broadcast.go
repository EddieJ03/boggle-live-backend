package main


func broadcastEndGame(room *Room, player1 float64, player2 float64) {
	// delete room first, then send endgame to clients
	clientRoomsLock.Lock() 

	delete(clientRooms, room.RoomName)
	
	clientRoomsLock.Unlock()

	room.Player1WS.Conn.WriteJSON(map[string]interface{}{
		"type": "endgame",
        "player1": player1,
    	"player2": player2,
	})

	room.Player2WS.Conn.WriteJSON(map[string]interface{}{
		"type": "endgame",
    	"player1": player1,
        "player2": player2,
	})
}

func broadcastDisconnect(roomName string) {
	clientRoomsLock.RLock()
	defer clientRoomsLock.RUnlock()

	room, exists := clientRooms[roomName]
	if !exists {
		return
	}

	room.Player1WS.Conn.WriteJSON(map[string]interface{}{
		"type":   "disconnected",
	})
	room.Player2WS.Conn.WriteJSON(map[string]interface{}{
		"type":   "disconnected",
	})
}

func broadcastSwitch(roomName string, player int, word string) {
	clientRoomsLock.RLock()
	defer clientRoomsLock.RUnlock()

	room, exists := clientRooms[roomName]
	if !exists {
		return
	}

	room.Player1WS.Conn.WriteJSON(map[string]interface{}{
		"type":   "switch",
		"player": player,
		"word": word,
	})
	room.Player2WS.Conn.WriteJSON(map[string]interface{}{
		"type":   "switch",
		"player": player,
		"word": word,
	})
}

func broadcastStart(roomName string) {
	clientRoomsLock.RLock()
	defer clientRoomsLock.RUnlock()

	room, exists := clientRooms[roomName]
	if !exists {
		return
	}

	room.Player1WS.Conn.WriteJSON(map[string]interface{}{
		"type":   "start",
		"countdown": [2]int{3,0},
		"gameInfo": *room,
	})
	room.Player2WS.Conn.WriteJSON(map[string]interface{}{
		"type":   "start",
		"countdown": [2]int{3,0},
		"gameInfo": *room,
	})
}