package main

import (
	"sync"
)

type Tile struct {
	I, J int
}

type Room struct {
	AllCharacters []string
	AllValidWords []string
	TotalScore    int
	Player1       float64 // score for player 1
	Player2       float64 // score for player 2
	Player1WS     *WSClient
	Player2WS 	  *WSClient
	Countdown     [2]int
	RoomLock      *sync.Mutex
	RoomName      string
	Player1MissedTurns int
	Player2MissedTurns int
}

type JoinGameMessage struct {
	Type string
	RoomName string
}

type NewGameMessage struct {
	Type string
}

type SubmitWordMessage struct {
	Type string
	Word string
	Score float64
}