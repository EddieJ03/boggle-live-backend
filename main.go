package main

import (
	"go_boggle_server/trie"
	"math/rand"
	"time"
	"fmt"

	"go_boggle_server/boards"
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
	Player1       int
	Player2       int
	TimerID       *time.Timer
	TimeOut       *time.Timer
}

var clientRooms = make(map[string]*Room)

var NUM = 4

var commonTrie = trie.NewTrie()

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

	clientRooms[roomName] = &Room{
		AllCharacters: allCharacters,
		AllValidWords: allValidWords,
		TotalScore:    totalScore,
		Player1:       0,
		Player2:       0,
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

func main() {
	common := boards.Common
	nursery := boards.Nursery
	shakespeare := boards.Shakespeare

	for key, value := range common {
        fmt.Printf("Key: %s, Value: %t\n", key, value)
    }

	for key, value := range nursery {
        fmt.Printf("Key: %s, Value: %t\n", key, value)
    }

	for key, value := range shakespeare {
        fmt.Printf("Key: %s, Value: %t\n", key, value)
    }
}

