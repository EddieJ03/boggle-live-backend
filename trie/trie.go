package trie

import "fmt"
import "strings"

// Node represents a node in the Trie
type Node struct {
    IsLast bool
    Next   [26]*Node
}

// Trie represents the trie data structure
type Trie struct {
    root   *Node
    OFFSET int
}

// NewTrie creates and returns a new Trie
func NewTrie() *Trie {
    return &Trie{
        root:   &Node{},
        OFFSET: 65,
    }
}

// Add adds a string to the Trie
func (t *Trie) Add(s string) {
    t.root = t.add2(t.root, s, 0)
}

// ContainsWord checks if the Trie contains the entire word
func (t *Trie) ContainsWord(s string) bool {
    upperS := strings.ToUpper(s)
    x := t.get(t.root, upperS, 0)
    if x == nil {
        return false
    }
    return x.IsLast
}

// ContainsPrefix checks if the Trie contains the prefix
func (t *Trie) ContainsPrefix(s string) bool {
    upperS := strings.ToUpper(s)

    return t.get(t.root, upperS, 0) != nil
}

// get retrieves a node for a given string
func (t *Trie) get(x *Node, s string, d int) *Node {
    if x == nil {
        return nil
    }
    if d == len(s) {
        return x
    }
    c := int(s[d]) - t.OFFSET

    if c > 25 {
        fmt.Printf("%c\n",s[d])
    }

    return t.get(x.Next[c], s, d+1)
}

// add2 adds a string to the Trie (helper function)
func (t *Trie) add2(x *Node, s string, d int) *Node {
    if x == nil {
        x = &Node{}
    }
    if d == len(s) {
        x.IsLast = true
        return x
    }
    c := int(s[d]) - t.OFFSET
    x.Next[c] = t.add2(x.Next[c], s, d+1)
    return x
}