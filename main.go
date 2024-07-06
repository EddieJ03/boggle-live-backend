package main

import (
	"encoding/json"
	"fmt"
	"go_boggle_server/boards"
	"go_boggle_server/trie"
	"math/rand"
	"net/http"
	"time"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

var BOGGLE_1992 = []string{
    "LRYTTE", "VTHRWE", "EGHWNE", "SEOTIS",
    "ANAEEG", "IDSYTT", "OATTOW", "MTOICU",
    "AFPKFS", "XLDERI", "HCPOAS", "ENSIEU",
    "YLDEVR", "ZNRNHL", "NMIQHU", "OBBAOJ",
}

// Define the 16 Boggle dice (1983 version)
var BOGGLE_1983 = []string{
    "AACIOT", "ABILTY", "ABJMOQ", "ACDEMP",
    "ACELRS", "ADENVZ", "AHMORS", "BIFORX",
    "DENOSW", "DKNOTU", "EEFHIY", "EGINTV",
    "EGKLUY", "EHINPS", "ELPSTU", "GILRUW",
}

// Define the 25 Boggle Master / Boggle Deluxe dice
var BOGGLE_MASTER = []string{
    "AAAFRS", "AAEEEE", "AAFIRS", "ADENNN", "AEEEEM",
    "AEEGMU", "AEGMNN", "AFIRSY", "BJKQXZ", "CCNSTW",
    "CEIILT", "CEILPT", "CEIPST", "DDLNOR", "DHHLOR",
    "DHHNOT", "DHLNOR", "EIIITT", "EMOTTT", "ENSSSU",
    "FIPRSY", "GORRVW", "HIPRRY", "NOOTUW", "OOOTTU",
}

// Define the 25 Big Boggle dice
var BOGGLE_BIG = []string{
    "AAAFRS", "AAEEEE", "AAFIRS", "ADENNN", "AEEEEM",
    "AEEGMU", "AEGMNN", "AFIRSY", "BJKQXZ", "CCENST",
    "CEIILT", "CEILPT", "CEIPST", "DDHNOT", "DHHLOR",
    "DHLNOR", "DHLNOR", "EIIITT", "EMOTTT", "ENSSSU",
    "FIPRSY", "GORRVW", "IPRRRY", "NOOTUW", "OOOTTU",
}

type Tile struct {
	I, J int
}

type Room struct {
	AllCharacters []string
	AllValidWords []string
	TotalScore    int
	Player1       int // score for player 1
	Player2       int // score for player 2
	TimerID       *time.Timer
	TimeOut       *time.Timer
	Player1WS     *WSClient
	Player2WS 	  *WSClient
	Countdown     [2]int
	RoomLock      *sync.Mutex
}

type WSClient struct {
	Conn     *websocket.Conn
	RoomName string
	Number   int
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
	Score int
}

var (
	clientRooms = make(map[string]*Room)
	clientRoomsLock sync.Mutex
)

const NUM = 4

var commonTrie = trie.NewTrie()

var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize:1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow all origins
    },
}

func initGame(roomName string) {
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

	allValidWords := findAllValidWords(constGrid)
	totalScore := calculateTotalPossibleScore(allValidWords)

	clientRoomsLock.Lock()
    defer clientRoomsLock.Unlock()

	clientRooms[roomName] = &Room{
		AllCharacters: allCharacters,
		AllValidWords: allValidWords,
		TotalScore:    totalScore,
		Player1:       0,
		Player2:       0,
		Player1WS:     nil,
		Player2WS:     nil,
		Countdown:	   [2]int{3,0},	
		RoomLock:      &sync.Mutex{},
	}
}


func findAllValidWords(constGrid [][]string) []string {
	words := []string{}

	for i := 0; i < NUM; i++ {
		for j := 0; j < NUM; j++ {
			newWords := dfs(i, j, constGrid)

			for _, word := range newWords {
				if !contains(words, word) {
					words = append(words, word)
				}
			}
		}
	}

	return words
}

func dfs(i, j int, constGrid [][]string) []string {
	s := Tile{i, j}

	marked := make([][]bool, NUM)
	for i := 0; i < NUM; i++ {
		marked[i] = make([]bool, NUM)
	}

	return dfs2(s, constGrid[i][j], marked, constGrid)
}

func dfs2(v Tile, prefix string, marked [][]bool, constGrid [][]string) []string {
	marked[v.I][v.J] = true

	words := []string{}

	if len(prefix) > 2 && commonTrie.ContainsWord(prefix) {
		words = append(words, prefix)
	}

	for _, adj := range adj2(v.I, v.J) {
		if !marked[adj.I][adj.J] {
			newWord := prefix + constGrid[adj.I][adj.J]
			if commonTrie.ContainsPrefix(newWord) {
				newWords := dfs2(adj, newWord, marked, constGrid)
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

// helper function to find if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	wsClient := &WSClient{Conn: conn}
	wsClient.HandleClient()
}

func (c *WSClient) HandleClient() {
	fmt.Println("IP:", c.Conn.RemoteAddr())

	c.Conn.SetCloseHandler(func(code int, text string) error {
		fmt.Println("Client disconnected")
		c.handleDisconnect()
		return nil
	})

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}

		var data map[string]interface{}
		err = json.Unmarshal(msg, &data)

		if err != nil {
			fmt.Println(err)
			continue
		}

		msgType, ok := data["type"].(string)
		if !ok {
			fmt.Println("Invalid type for message Type")
			continue
		}

		fmt.Println("messageType: " + msgType)

		switch msgType {
		case "newGame":
			c.newGame()
		case "joinGame":
			var data JoinGameMessage
			err = json.Unmarshal(msg, &data)

			if err != nil {
				fmt.Println(err)
				continue
			}

			c.joinGame(data.RoomName)
		case "submitWord":
			var data SubmitWordMessage
			err = json.Unmarshal(msg, &data)

			if err != nil {
				fmt.Println(err)
				continue
			}
			
			c.submitWord(data)
		default:
			continue
		}
	}
}

func (c *WSClient) newGame() {
	for item := range boards.Common {
		commonTrie.Add(item)
	}

	roomName := makeID(15)

	c.Conn.WriteJSON(map[string]string{
		"type":     "gameCode",
		"roomName": roomName,
	})

	c.RoomName = roomName
	c.Number = 1

	initGame(roomName)

	room, exists := clientRooms[roomName]
	if !exists {
		return
	}

	room.Player1WS = c

	c.Conn.WriteJSON(map[string]interface{}{
		"type":   "init",
		"number": 1,
	})
}

func (c *WSClient) joinGame(roomName string) {
	numClients := numberOfClients(roomName)

	if numClients == 0 {
		c.Conn.WriteJSON(map[string]string{
			"type": "unknownGame",
		})
		return
	} else if numClients > 1 {
		c.Conn.WriteJSON(map[string]string{
			"type": "tooManyPlayers",
		})
		return
	}

	c.RoomName = roomName
	c.Number = numClients + 1

	c.Conn.WriteJSON(map[string]interface{}{
		"type":   "init",
		"number": numClients + 1,
	})

	room, exists := clientRooms[roomName]
	if !exists {
		return
	}

	room.Player2WS = c

	startGame(roomName)
}

func (c *WSClient) submitWord(data SubmitWordMessage) {
	room, exists := clientRooms[c.RoomName]
	if !exists {
		return
	}

	room.RoomLock.Lock()
    defer room.RoomLock.Unlock()

    if c.Number == 1 {
        fmt.Println("Switching to player 2")
        room.Player1 = data.Score
        broadcastSwitch(c.RoomName, 2, data.Word)
    } else {
        fmt.Println("Switching to player 1")
        room.Player2 = data.Score
        broadcastSwitch(c.RoomName, 1, data.Word)
    }
}

func (c *WSClient) handleDisconnect() {
	if c.RoomName == "" {
		fmt.Println("could not find room to delete after disconnect!")
		return
	}

	broadcastDisconnect(c.RoomName)

	_, exists := clientRooms[c.RoomName]
	if !exists {
		fmt.Println("could not find room to delete after disconnect!")
		return
	}

	delete(clientRooms, c.RoomName);
}

// used in joinGame
func startGame(roomName string) {
	minute := 3;
	seconds := 0;
	countdown := 180;

	_, exists := clientRooms[roomName]
	if !exists {
		return
	}

	broadcastStart(roomName);

	ticker := time.NewTicker(1 * time.Second)
    timer := time.NewTimer(time.Duration(countdown) * time.Second)

    go func() {
        for {
            select {
            case <-ticker.C:
				// terminate goroutine if room is gone
				room, exists := clientRooms[roomName]
				if !exists {
					return
				}

                broadcastTime(room, minute, seconds)

                if seconds == 0 {
                    if minute == 0 {
                        ticker.Stop()
                        return
                    }

                    seconds = 59
                    minute--
                } else {
                    seconds--
                }
            case <-timer.C:
                ticker.Stop()
                gameData := clientRooms[roomName]

                broadcastEndGame(roomName, gameData.Player1, gameData.Player2)

                delete(clientRooms, roomName)
                return
            }
        }
    }()

    // Keep the main function alive to allow the goroutines to run
    select {}
}

// used in joinGame
func numberOfClients(roomName string) int {
	room, exists := clientRooms[roomName]
	if !exists {
		return 0
	}

	num := 0
	if room.Player1WS != nil {
		num++
	}

	if room.Player2WS != nil {
		num++
	}

	return num
}

// used in newGame
func makeID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)

	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	
	return string(b)
}

func broadcastEndGame(roomName string, player1 int, player2 int) {
	room, exists := clientRooms[roomName]
	if !exists {
		return
	}

	err := room.Player1WS.Conn.WriteJSON(map[string]interface{}{
		"type": "endgame",
        "player1": player1,
    	"player2": player2,
	})

	if err != nil {
		fmt.Println("Error writing message:", err)
		// Optionally close the connection on error
		return
	}

	err2 := room.Player2WS.Conn.WriteJSON(map[string]interface{}{
		"type": "endgame",
    	"player1": player1,
        "player2": player2,
	})

	if err2 != nil {
		fmt.Println("Error writing message:", err2)
		// Optionally close the connection on error
		return
	}
}

func broadcastTime(room *Room, minute int, seconds int) {
	room.Player1WS.Conn.WriteJSON(map[string]interface{}{
		"type": "time",
        "time": []int{minute, seconds},
	})
	room.Player2WS.Conn.WriteJSON(map[string]interface{}{
		"type": "time",
    	"time": []int{minute, seconds},
	})
}

func broadcastDisconnect(roomName string) {
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

