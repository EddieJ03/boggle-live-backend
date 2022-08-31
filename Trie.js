// Node Class
class Node {
    constructor() {
      this.isLast = false
      this.next = new Array(26)
  
      for(let i = 0; i < this.next.length; i++) {
        this.next[i] = null
      }
    }
}
  
// Trie Class
class Trie {
    constructor() {
      this.root = new Node()
      this.OFFSET = 65
    }
  
    // CANNOT HAVE DUPLICATE METHOD NAMES IN JAVASCRIPT - 
    // LAST ONE PARSED WILL REPLACE PREVIOUS ONE (known as hoisting)
    add(s) {
        // update root after adding
        this.root = this.add2(this.root, s, 0);
    }
  
    // isLast is a boolean indicating if end of string is reached
    containsWord(s) {
        let x = this.get(this.root, s, 0);
  
        if (x === null) {
            return false;
        }
  
        // string s only contained if Node got from get is a leaf node
        return x.isLast;
    }
  
    containsPrefix(s) {
        let x = this.get(this.root, s, 0);
  
        return x != null;
    }
  
    get(x, s, d) {
        // base case
        if (x == null) {
            return null;
        }
  
        // reached end of string s, return Node currently at
        if (d === s.length) {
            return x;
        }
  
        // so we can index properly into the array
        let c = s.charCodeAt(d) - this.OFFSET;
        
        // update counter d
        return this.get(x.next[c], s, d + 1);
    }
  
  
    // method that does adding to Trie
    add2(x, s, d) {
      // instantiate x if reached null spot in array
        if (x === null) {
            x = new Node();
        }
  
        if (d === s.length) {
            x.isLast = true;
            return x;
        }
  
        // account for OFFSET since array is labeled with indices 0 to 25
        let c = s.charCodeAt(d) - this.OFFSET;
  
        // update Node array at index c
        x.next[c] = this.add2(x.next[c], s, d + 1);
  
        // return updated Node
        return x;
    }
}

module.exports = Trie