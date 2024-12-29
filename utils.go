package main

import (
	"context"
	"go_boggle_server/trie"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

// change this endpoint depending on where server is (ex: localhost:9094)
var endpoint string = "52.160.91.109:9094"

func startGame(room *Room) {
	roomName := room.RoomName
	broadcastStart(roomName)
}


func findAllValidWords(constGrid [][]string, trie *trie.Trie) []string {
	words := []string{}

	for i := 0; i < NUM; i++ {
		for j := 0; j < NUM; j++ {
			newWords := dfs(i, j, constGrid, trie)

			for _, word := range newWords {
				if !contains(words, word) {
					words = append(words, word)
				}
			}
		}
	}

	return words
}

func initGame(roomName string, trie *trie.Trie, random bool) {
	constGrid := make([][]string, NUM)
	allCharacters := []string{}

	for i := 0; i < NUM; i++ {
		constGrid[i] = []string{}
	}

	chosenBoggle := BOGGLE_1992
	if rand.Intn(2) == 0 {
		chosenBoggle = BOGGLE_1983
	}

	for i := 0; i < NUM*NUM; i++ {
		randIndex := rand.Intn(6)
		char := chosenBoggle[i][randIndex : randIndex+1]
		if char == "Q" {
			char += "u"
		}
		constGrid[i/NUM] = append(constGrid[i/NUM], char)
		allCharacters = append(allCharacters, char)
	}

	allValidWords := findAllValidWords(constGrid, trie)
	totalScore := calculateTotalPossibleScore(allValidWords)

	clientRoomsLock.Lock()
    defer clientRoomsLock.Unlock()
	
	room := &Room{
		AllCharacters: allCharacters,
		AllValidWords: allValidWords,
		TotalScore:    totalScore,
		Player1:       0,
		Player2:       0,
		Player1WS:     nil,
		Player2WS:     nil,
		Countdown:	   [2]int{3,0},	
		RoomLock:      &sync.Mutex{},
		RoomName:      roomName,
		Player1MissedTurns: 0,
		Player2MissedTurns: 0,
	}

	_, topic_err := kafka.DialLeader(context.Background(), "tcp", endpoint, roomName, 0) // this creates topic since the kafka config is set to auto topic creation
	if topic_err != nil {
		room.KafkaWriter = nil
		log.Printf("failed to create topic: %s\n", topic_err.Error())
	} else {
		room.KafkaWriter = &kafka.Writer{
			Addr:     kafka.TCP(endpoint),
			Topic:   roomName,
			RequiredAcks: kafka.RequireAll,
			Async:        true,
			BatchSize:    1,
		}
	}

	clientRooms[roomName] = room

	// fmt.Println("successfully created room!")

	// if random, also add to the list
	if(random) {
		// fmt.Println("added to random!")
		randomRooms = append(randomRooms, room)
	}
}

func dfs(i, j int, constGrid [][]string, trie *trie.Trie) []string {
	s := Tile{i, j}

	marked := make([][]bool, NUM)
	for i := 0; i < NUM; i++ {
		marked[i] = make([]bool, NUM)
	}

	return dfs2(s, constGrid[i][j], marked, constGrid, trie)
}

func dfs2(v Tile, prefix string, marked [][]bool, constGrid [][]string, commonTrie *trie.Trie) []string {
	marked[v.I][v.J] = true

	words := []string{}

	if len(prefix) > 2 && commonTrie.ContainsWord(prefix) {
		words = append(words, prefix)
	}

	for _, adj := range adj2(v.I, v.J) {
		if !marked[adj.I][adj.J] {
			newWord := prefix + constGrid[adj.I][adj.J]
			if commonTrie.ContainsPrefix(newWord) {
				newWords := dfs2(adj, newWord, marked, constGrid, commonTrie)
				words = append(words, newWords...)
			}
		}
	}

	marked[v.I][v.J] = false
	return words
}

func adj2(i, j int) []Tile {
	adj := []Tile{}

	if i > 0 {
		adj = append(adj, Tile{i - 1, j})
		if j > 0 {
			adj = append(adj, Tile{i - 1, j - 1})
		}
		if j < NUM-1 {
			adj = append(adj, Tile{i - 1, j + 1})
		}
	}

	if i < NUM-1 {
		adj = append(adj, Tile{i + 1, j})
		if j > 0 {
			adj = append(adj, Tile{i + 1, j - 1})
		}
		if j < NUM-1 {
			adj = append(adj, Tile{i + 1, j + 1})
		}
	}

	if j > 0 {
		adj = append(adj, Tile{i, j - 1})
	}
	if j < NUM-1 {
		adj = append(adj, Tile{i, j + 1})
	}

	return adj
}

func calculateTotalPossibleScore(allValidWords []string) int {
	total := 0

	for _, word := range allValidWords {
		switch len(word) {
		case 3, 4:
			total++
		case 5:
			total += 2
		case 6:
			total += 3
		case 7:
			total += 5
		default:
			total += 11
		}
	}

	return total
}

func makeID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)

	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	
	return string(b)
}

func numberOfClients(room *Room) int {
	num := 0
	if room.Player1WS != nil {
		num++
	}

	if room.Player2WS != nil {
		num++
	}

	return num
}

// helper function to find if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func findRoomIndex(rooms []*Room, name string) int {
    for i, room := range rooms {
        if room.RoomName == name {
            return i
        }
    }

    return -1 // return -1 if not found
}

// Function to remove a Room from the list by index
func removeRoom(rooms []*Room, index int) []*Room {
    if index < 0 || index >= len(rooms) {
        return rooms // return the original slice if index is out of bounds
    }

    return append(rooms[:index], rooms[index+1:]...)
}

func popFirstRoom(rooms []*Room) ([]*Room, *Room) {
    if len(rooms) == 0 {
        return rooms, nil // return the original slice and nil if it's empty
    }
    firstRoom := rooms[0]
    return rooms[1:], firstRoom
}

func sendMessage(room *Room, message string) {
	if room.KafkaWriter == nil {
		return
	}

	room.KafkaWriter.WriteMessages(
		context.Background(),
		kafka.Message{
			Value: []byte(message),
		},
	)
}

func deleteTopic(topic string) {
	conn, err := kafka.Dial("tcp", endpoint)

	if err != nil {
		log.Println("failed to dial to remove topic " + topic)
		return
	}

	defer conn.Close()

	conn.DeleteTopics(topic)
}
