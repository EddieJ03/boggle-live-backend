// server.js
const Trie = require('./Trie.js');
const COMMON = require('./common.js');
const { BOGGLE_1992, BOGGLE_1983 } = require('./boards.js');

const express = require('express');
const app = express();
const http = require('http');
const WebSocket = require('ws');

app.use(require('cors')());

const server = http.createServer(app);
const wss = new WebSocket.Server({ server });

const NUM = 4;

// map room to array of characters, all valid words, and total possible score
const clientRooms = {};
const commonTrie = new Trie();

class Tile {
    constructor(iVal, jVal) {
        this.i = iVal;
        this.j = jVal;
    }
}

wss.on('connection', ws => {
    console.log("IP: ", ws._socket.remoteAddress);

    ws.on('close', () => {
        console.log('Client disconnected');
        handleDisconnect(ws);
    });

    ws.on('message', message => {
        const data = JSON.parse(message);

        switch (data.type) {
            case 'newGame':
                newGame(ws);
                break;
            case 'joinGame':
                joinGame(ws, data.roomName);
                break;
            case 'submitWord':
                submitWord(ws, data);
                break;
            default:
                break;
        }
    });

    function numberOfClients(roomName) {
        const clients = Array.from(wss.clients).filter(client => client.roomName === roomName);
        return clients.length;
    }

    function joinGame(ws, roomName) {    
        let numClients = numberOfClients(roomName);

        if (numClients === 0) {
            ws.send(JSON.stringify({ type: 'unknownCode' }));
            return;
        } else if (numClients > 1) {
            ws.send(JSON.stringify({ type: 'tooManyPlayers' }));
            return;
        }

        clientRooms[ws.id] = roomName;

        ws.roomName = roomName;
        ws.number = 2;

        ws.send(JSON.stringify({ type: 'init', number: 2 }));
        startGame(roomName);
    }

    function makeid(length) {
        var result           = '';
        var characters       = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
        var charactersLength = characters.length;
        for ( var i = 0; i < length; i++ ) {
           result += characters.charAt(Math.floor(Math.random() * charactersLength));
        }
        return result;
    }

    function newGame(ws) {
        for(const item in COMMON) {
            commonTrie.add(`${item}`);
        }

        let roomName = makeid(15);

        ws.send(JSON.stringify({ type: 'gameCode', roomName }));

        ws.roomName = roomName;
        ws.number = 1;

        initGame(roomName);

        ws.send(JSON.stringify({ type: 'init', number: 1 }));
    }

    function initGame(roomName) {
        for(const item in COMMON) {
            commonTrie.add(`${item}`);
        }

        let constGrid = [], allCharacters = [];

        for(let i = 0; i < NUM; i++) {
            constGrid.push(new Array());
        }

        let chosenBoggle = Math.floor(Math.random() * 2) === 0 ? BOGGLE_1992 : BOGGLE_1983;

        for (let i = 0; i < NUM * NUM; i++) {
            let rand = Math.floor(Math.random() * 6);
            let char = chosenBoggle[i].substring(rand, rand + 1);
            char = char === 'Q' ? char + 'u' : char;
            constGrid[Math.floor(i / 4)].push(char);
            allCharacters.push(char);
        }

        let allValidWords = findAllValidWords(constGrid);

        let totalScore = calculateTotalPossibleScore(allValidWords);

        clientRooms[roomName] = {
            allCharacters: allCharacters,
            allValidWords: allValidWords,
            totalScore: totalScore,
            player1: 0,
            player2: 0,
        };
    }

    function calculateTotalPossibleScore(allValidWords) {
        let total = 0;
    
        for(let i = 0; i < allValidWords.length; i++) {
          if(allValidWords[i].length >= 3 && allValidWords[i].length <= 4) total += 1;
          else if(allValidWords[i].length === 5) total += 2;
          else if(allValidWords[i].length === 6) total += 3;
          else if(allValidWords[i].length === 7) total += 5;
          else total += 11;
        }
    
        return total;
    }

    function findAllValidWords(constGrid) {
        let words = [];
    
        // loop through entire grid and find words starting at that cell
        for (let i = 0; i < NUM; ++i) {
          for (let j = 0; j < NUM; ++j) {
            let newWords = dfs(i, j, constGrid);
            // newWords returns an array; only add words already NOT in words array to words array
            for(let i = 0; i < newWords.length; i++) {
              if(!words.includes(newWords[i])) {
                words.push(newWords[i]);
              }
            }
          }
        }
        
        return words;
    }

    function dfs(i, j, constGrid) {
        let s = new Tile(i, j);
        
        let marked = [];
    
        // set everything to false in 2D array to false: 
        // supposed to reflect visited spots on grid
        for(let i = 0; i < NUM; i++) {
          marked.push(new Array(NUM));
          for(let j = 0; j < NUM; j++) {
            marked[i].push(false);
          }
        }
    
        // run dfs2 starting from this cell
        return dfs2(s, "" + constGrid[i][j], marked, constGrid);
    }
    
    function dfs2(v, prefix, marked, constGrid) {
        // visited current spot at v.i and v.j; mark as true
        marked[v.i][v.j] = true;
        
        let words = [];
    
        // word length ust be at least two AND if it exists in
        // Trie, add to words array
        if (prefix.length > 2 && commonTrie.containsWord(prefix)) {
            words.push(prefix);
        }
    
        // get all adjacent cells
        let arr = adj2(v.i, v.j);
    
        // loop through every cell and if NOT visited yet, update
        // prefix by adding character and call dfs2 AGAIN
        for (let z = 0; z < arr.length; z++) {
    
            if (!marked[arr[z].i][arr[z].j]) {
    
                // newWord is updated prefix
                let newWord = prefix + constGrid[arr[z].i][arr[z].j];
    
                if (commonTrie.containsPrefix(newWord)) {
                  // continue getting words starting at newWord
                  let newWords = dfs2(arr[z], newWord, marked, constGrid)
                  
                  // update words with all found words in newWords
                  words = [...words, ...newWords]
                }
    
            }
        }
    
        // when finished set spot to FALSE so other backtracking calls can visit
        marked[v.i][v.j] = false;
    
        // all possible words
        return words;
    }

    function adj2(i, j) {
        let adj = [];
  
        if (i > 0) {
            adj.push(new Tile(i - 1, j));
            if (j > 0) adj.push(new Tile(i - 1, j - 1));
            if (j < NUM - 1) adj.push(new Tile(i - 1, j + 1));
        }
  
        if (i < NUM - 1) {
            adj.push(new Tile(i + 1, j));
            if (j > 0) adj.push(new Tile(i + 1, j - 1));
            if (j < NUM - 1) adj.push(new Tile(i + 1, j + 1));
        }
  
        if (j > 0) adj.push(new Tile(i, j - 1));
        if (j < NUM - 1) adj.push(new Tile(i, j + 1));
  
        return adj;
  }

    function startGame(roomName) {
        let minute = 2;
        let seconds = 59;

        let countdown = 181;

        let room = clientRooms[roomName];

        broadcast(roomName, {
            type: 'start',
            countdown: [3, 0],
            gameInfo: clientRooms[roomName]
        });

        let timerId = setInterval(() => {
            broadcast(roomName, {
                type: 'time',
                time: [minute, seconds]
            });
            if(seconds === 0) {
                seconds = 59;
                minute -= 1;
            } else {
                seconds -= 1;
            }
        }, 1000);

        let timeOut = setTimeout(() => {
            clearInterval(timerId);
            let gameData = clientRooms[ws.roomName];
            broadcast(ws.roomName, {
                type: 'endgame',
                player1: gameData.player1,
                player2: gameData.player2
            });
            delete clientRooms[roomName];
        }, countdown * 1000);

        room.timerId = timerId;
        room.timeOut = timeOut;
    }

    function submitWord(ws, data) {
        let room = clientRooms[ws.roomName];

        if(ws.number === 1) {
            room.player1 = data.score;
            broadcast(ws.roomName, {
                type: 'switch',
                player: 2,
                word: data.word
            });
        } else {
            room.player2 = data.score;
            broadcast(ws.roomName, {
                type: 'switch',
                player: 1,
                word: data.word
            });
        }
    }

    function handleDisconnect(ws) {
        if(ws.roomName) {
            let gameData = clientRooms[ws.roomName];
            if(gameData) {
                clearInterval(gameData.timerId);
                clearTimeout(gameData.timeOut);
                broadcast(ws.roomName, {
                    type: 'disconnected'
                });
                delete clientRooms[ws.roomName];
            }
        }
    }

    function broadcast(roomName, data) {
        wss.clients.forEach(client => {
            if(client.readyState === WebSocket.OPEN && client.roomName === roomName) {
                client.send(JSON.stringify(data));
            }
        });
    }
});

server.listen(process.env.PORT || 5000, () => {
    console.log("Server is running on port 5000!");
});
