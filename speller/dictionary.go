// Package dictionary implements a dictionary's functionality
package dictionary

import (
	"bufio"
	"os"
	"strings"
	"unicode"
)

// ── internal state ──────────────────────────────────────────────────────────

// Node represents a node in the hash table.
// In Go, the garbage collector automatically handles freeing these when they
// are no longer referenced.
type Node struct {
	word string
	next *Node
}

// N is the number of buckets in the hash table
const N = 28

// table is the hash table: each bucket points to the head of a singly-linked chain.
// Lowercase name keeps it unexported (internal to the package).
var table [N]*Node

// wordCount tracks the number of words currently loaded
var wordCount uint = 0

// ── public API ──────────────────────────────────────────────────────────────

// Hash hashes a word to a bucket index.
func Hash(word string) uint {
	var sum uint = 0
	// Iterating over the length acts on bytes, matching C++'s string char iteration
	for i := 0; i < len(word); i++ {
		sum += uint(unicode.ToLower(rune(word[i])))
	}
	return sum % N
}

// Check returns true if word is in the dictionary, else false.
// Comparison is case-insensitive.
func Check(word string) bool {
	bucket := Hash(word)
	for ptr := table[bucket]; ptr != nil; ptr = ptr.next {
		// EqualFold performs a case-insensitive string comparison
		if strings.EqualFold(word, ptr.word) {
			return true
		}
	}
	return false
}

// Load loads dictionary into memory; returns true on success, false on failure.
func Load(dictionary string) bool {
	file, err := os.Open(dictionary)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// ScanWords mimics the behavior of C++'s `while (file >> word)`
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		word := scanner.Text()

		n := &Node{
			word: word,
		}

		bucket := Hash(word)
		// Prepend to chain
		n.next = table[bucket]
		table[bucket] = n

		wordCount++
	}

	// Check for any read errors during scanning
	if err := scanner.Err(); err != nil {
		return false
	}

	return true
}

// Size returns number of words loaded, or 0 if not yet loaded.
func Size() uint {
	return wordCount
}

// Unload unloads the dictionary from memory; returns true.
// Dropping the pointers makes the entire linked list eligible for garbage collection.
func Unload() bool {
	for i := range table {
		table[i] = nil
	}
	wordCount = 0
	return true
}
