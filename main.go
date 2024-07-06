package main

import (
	// "encoding/json"
	"fmt"
	"go_boggle_server/boards"
	"go_boggle_server/trie"
	"math/rand"
	"net/http"
	"sync"
	"time"
	"unicode/utf8"

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

type WSClient struct {
	Conn           *websocket.Conn
	RoomName       string
	UniqueNumber   int
	Number         int
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

var (
	clientRooms = make(map[string]*Room)
	clientRoomsLock sync.RWMutex
)

const NUM = 4

var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize:1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow all origins
    },
}

func initGame(roomName string, trie *trie.Trie) {
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
		RoomName:      roomName,
		Player1MissedTurns: 0,
		Player2MissedTurns: 0,
	}
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

	wsClient := &WSClient{Conn: conn, RoomName: "", UniqueNumber: rand.Int(), Number: -1}
	wsClient.HandleClient()
}

func (c *WSClient) HandleClient() {
	fmt.Printf("%d connected\n", c.UniqueNumber)

	c.Conn.SetCloseHandler(func(code int, text string) error {
		c.handleDisconnect()
		fmt.Printf("%d closed so disconnected\n", c.UniqueNumber)
		return nil
	})

	for {
		var data map[string]interface{}

		err := c.Conn.ReadJSON(&data)
		if err != nil {
			fmt.Println(err)

			// disconnect both when error
			c.handleDisconnect()

			fmt.Printf("%d error so disconnected\n", c.UniqueNumber)

			break
		}

		msgType, ok := data["type"].(string)
		if !ok {
			fmt.Printf("%s is nvalid type for message Type\n", msgType)
			continue
		}

		fmt.Println("messageType: " + msgType)

		switch msgType {
		case "newGame":
			c.newGame()
		case "joinGame":
			c.joinGame(data["roomName"].(string))
		case "submitWord":
			swm := SubmitWordMessage{
				Type: data["type"].(string),
				Word: data["word"].(string),
				Score: data["score"].(float64),
			}

			c.submitWord(swm)
		default:
			continue
		}
	}
}

func (c *WSClient) newGame() {
	var commonTrie = trie.NewTrie()

	for item := range boards.Common {
		commonTrie.Add(item)
	}

	roomName := makeID(15)

	c.RoomName = roomName
	c.Number = 1

	c.Conn.WriteJSON(map[string]string{
		"type":     "gameCode",
		"roomName": roomName,
	})

	initGame(roomName, commonTrie)

	clientRoomsLock.RLock()
	defer clientRoomsLock.RUnlock()

	room, exists := clientRooms[roomName]
	if !exists {
		return
	}

	room.Player1WS = c

	c.Conn.WriteJSON(map[string]interface{}{
		"type":   "init",
		"number": 1,
	})

	fmt.Printf("%d is player %d in room %s\n", c.UniqueNumber, c.Number, c.RoomName)
}

func (c *WSClient) joinGame(roomName string) {
	clientRoomsLock.RLock()
	defer clientRoomsLock.RUnlock()

	room, exists := clientRooms[roomName]
	if !exists {
		fmt.Println("Room " + roomName + " does not exist!")
		return
	}

	// one new player at a time should be here
	room.RoomLock.Lock()

	numClients := numberOfClients(room)

	if numClients == 0 || numClients == -1 {
		fmt.Println("Room " + roomName + " has 0 players??!")
		c.Conn.WriteJSON(map[string]string{
			"type": "unknownGame",
		})
		return
	} else if numClients > 1 {
		fmt.Println("Room " + roomName + " has too many players??!")
		c.Conn.WriteJSON(map[string]string{
			"type": "tooManyPlayers",
		})
		return
	}

	c.Number = 2
	c.RoomName = roomName
	room.Player2WS = c

	c.Conn.WriteJSON(map[string]interface{}{
		"type":   "init",
		"number": 2,
	})

	fmt.Printf("%d is player %d in room %s\n", c.UniqueNumber, c.Number, c.RoomName)
	
	room.RoomLock.Unlock()

	startGame(room)
}

func (c *WSClient) submitWord(data SubmitWordMessage) {
	clientRoomsLock.RLock()
	defer clientRoomsLock.RUnlock()

	room, exists := clientRooms[c.RoomName]
	if !exists {
		return
	}

    if c.Number == 1 {
        fmt.Println("Switching to player 2")

		if utf8.RuneCountInString(data.Word) == 0 {
			room.Player1MissedTurns += 1
		} else {
			room.Player1MissedTurns = 0
		}

		if room.Player1MissedTurns == 3 {
			broadcastEndGame(c.RoomName, room.Player1, room.Player2)
		}

        room.Player1 = data.Score
        broadcastSwitch(c.RoomName, 2, data.Word)
    } else {
        fmt.Println("Switching to player 1")

		if utf8.RuneCountInString(data.Word) == 0 {
			room.Player2MissedTurns += 1
		} else {
			room.Player2MissedTurns = 0
		}

		if room.Player2MissedTurns == 3 {
			broadcastEndGame(c.RoomName, room.Player1, room.Player2)
		}

        room.Player2 = data.Score
        broadcastSwitch(c.RoomName, 1, data.Word)
    }
}

func (c *WSClient) handleDisconnect() {
	if c.RoomName == "" {
		fmt.Printf("%d could not find room %s to delete after disconnect!\n", c.Number, c.RoomName)
		return
	}

	broadcastDisconnect(c.RoomName)
	
	clientRoomsLock.Lock()	

	delete(clientRooms, c.RoomName)

	clientRoomsLock.Unlock()

	fmt.Printf("%d found room %s to delete after disconnect!\n", c.Number, c.RoomName)
}

// used in joinGame
func startGame(room *Room) {
	// minute := 3;
	// seconds := 0;
	// countdown := 181;

	roomName := room.RoomName

	broadcastStart(roomName)
}

// used in joinGame
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

func broadcastEndGame(roomName string, player1 float64, player2 float64) {
	clientRoomsLock.RLock() 

	room, exists := clientRooms[roomName]
	if !exists {
		return
	}

	clientRoomsLock.RUnlock()

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

