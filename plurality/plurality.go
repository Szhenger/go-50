package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Max number of candidates
const MAX = 9

// Candidates have name and vote count.
type Candidate struct {
	name  string
	votes int
}

// Array of candidates
var candidates [MAX]Candidate

// Number of candidates
var candidateCount int

func main() {
	// Check for invalid usage
	if len(os.Args) < 2 {
		fmt.Println("Usage: plurality [candidate ...]")
		os.Exit(1)
	}

	// Populate array of candidates
	candidateCount = len(os.Args) - 1
	if candidateCount > MAX {
		fmt.Printf("Maximum number of candidates is %d\n", MAX)
		os.Exit(2)
	}

	for i := 0; i < candidateCount; i++ {
		candidates[i].name = os.Args[i+1]
		// In Go, struct fields are automatically initialized to their "zero value".
		// For integers, this is 0, so we don't need to manually set it.
	}

	// Create a single scanner for all standard input
	scanner := bufio.NewScanner(os.Stdin)

	voterCount := getVoterCount(scanner)

	// Loop over all voters
	for i := 0; i < voterCount; i++ {
		fmt.Print("Vote: ")

		// Read the entire line and clear leftover whitespace/newlines
		scanner.Scan()
		name := strings.TrimSpace(scanner.Text())

		// Check for invalid vote
		if !vote(name) {
			fmt.Println("Invalid vote.")
		}
	}

	// Display winner of election
	printWinner()
}

// Safely grabs an integer for voter count
func getVoterCount(scanner *bufio.Scanner) int {
	for {
		fmt.Print("Number of voters: ")

		scanner.Scan()
		input := scanner.Text()

		count, err := strconv.Atoi(strings.TrimSpace(input))
		if err == nil && count > 0 {
			return count
		}
	}
}

// Update vote totals given a new vote
func vote(name string) bool {
	for i := 0; i < candidateCount; i++ {
		// Go strings can be directly compared using ==
		if name == candidates[i].name {
			candidates[i].votes++
			return true
		}
	}
	return false
}

// Print the winner (or winners) of the election
func printWinner() {
	maxVotes := 0

	// Step 1: Find the maximum number of votes
	for i := 0; i < candidateCount; i++ {
		if candidates[i].votes > maxVotes {
			maxVotes = candidates[i].votes
		}
	}

	// Step 2: Print the name of any candidate who has that max amount
	for i := 0; i < candidateCount; i++ {
		if candidates[i].votes == maxVotes {
			fmt.Println(candidates[i].name)
		}
	}
}
