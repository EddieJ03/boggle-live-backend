package main

import "fmt"

func broadcastEndGame(room *Room, player1 float64, player2 float64) {
	// delete room first, then send endgame to clients
	clientRoomsLock.Lock()

	delete(clientRooms, room.RoomName)

	clientRoomsLock.Unlock()

	room.Player1WS.Conn.WriteJSON(map[string]interface{}{
		"type":    "endgame",
		"player1": player1,
		"player2": player2,
	})

	room.Player2WS.Conn.WriteJSON(map[string]interface{}{
		"type":    "endgame",
		"player1": player1,
		"player2": player2,
	})

	gameOverMessage := "GAME OVER!"

	if room.Player1MissedTurns == 3 {
		gameOverMessage = gameOverMessage + "\n Player 1 Missed 3 Consecutive Turns"
	} else if room.Player2MissedTurns == 3 {
		gameOverMessage = gameOverMessage + "\n Player 2 Missed 3 Consecutive Turns"
	} else {
		gameOverMessage = gameOverMessage + "\n All possible words found!"
	}

	sendMessage(room, gameOverMessage)
	deleteTopic(room.RoomName)
}

func broadcastDisconnect(roomName string) {
	clientRoomsLock.RLock()
	defer clientRoomsLock.RUnlock()

	room, exists := clientRooms[roomName]
	if !exists {
		return
	}

	room.Player1WS.Conn.WriteJSON(map[string]interface{}{
		"type": "disconnected",
	})
	room.Player2WS.Conn.WriteJSON(map[string]interface{}{
		"type": "disconnected",
	})

	sendMessage(room, "Someone disconnected! This game has ended.")
	deleteTopic(room.RoomName)
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
		"word":   word,
	})
	room.Player2WS.Conn.WriteJSON(map[string]interface{}{
		"type":   "switch",
		"player": player,
		"word":   word,
	})

	sendMessage(room, fmt.Sprintf("Player %d found word %s. Switching from Player %d to %d", 1-player, word, 1-player, player))
}

func broadcastStart(roomName string) {
	clientRoomsLock.RLock()
	defer clientRoomsLock.RUnlock()

	room, exists := clientRooms[roomName]
	if !exists {
		return
	}

	room.Player1WS.Conn.WriteJSON(map[string]interface{}{
		"type":      "start",
		"countdown": [2]int{3, 0},
		"gameInfo":  *room,
	})
	room.Player2WS.Conn.WriteJSON(map[string]interface{}{
		"type":      "start",
		"countdown": [2]int{3, 0},
		"gameInfo":  *room,
	})

	sendMessage(room, "Game start!")
}