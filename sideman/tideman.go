package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Max number of candidates
const MAX = 9

// preferences[i][j] is number of voters who prefer i over j
// Global arrays in Go are automatically initialized to 0 / false.
var preferences [MAX][MAX]int

// locked[i][j] means i is locked in over j
var locked [MAX][MAX]bool

// Each pair has a winner, loser.
type Pair struct {
	winner int
	loser  int
}

// Array of candidates and pairs
var candidates [MAX]string
var pairs [(MAX * (MAX - 1)) / 2]Pair

// Helper variables
var pairCount int
var candidateCount int

func main() {
	// Check for invalid usage
	if len(os.Args) < 2 {
		fmt.Println("Usage: tideman [candidate ...]")
		os.Exit(1)
	}

	// Populate array of candidates
	candidateCount = len(os.Args) - 1
	if candidateCount > MAX {
		fmt.Printf("Maximum number of candidates is %d\n", MAX)
		os.Exit(2)
	}
	for i := 0; i < candidateCount; i++ {
		candidates[i] = os.Args[i+1]
	}

	// Create a single scanner to handle all standard input safely
	scanner := bufio.NewScanner(os.Stdin)

	voterCount := getVoterCount(scanner)

	// Query for votes
	for i := 0; i < voterCount; i++ {
		// Using a slice created via make() instead of std::vector
		ranks := make([]int, candidateCount)

		// Query for each rank
		for j := 0; j < candidateCount; j++ {
			fmt.Printf("Rank %d: ", j+1)

			scanner.Scan()
			name := strings.TrimSpace(scanner.Text())

			// Slices are passed by reference natively in Go
			if !vote(j, name, ranks) {
				fmt.Println("Invalid vote.")
				os.Exit(3)
			}
		}

		recordPreferences(ranks)
		fmt.Println() // Prints a blank line
	}

	addPairs()
	sortPairs()
	lockPairs()
	printWinner()
}

// Safely grabs an integer
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

// Update ranks given a new vote
func vote(rank int, name string, ranks []int) bool {
	for i := 0; i < candidateCount; i++ {
		if name == candidates[i] {
			ranks[rank] = i
			return true
		}
	}
	return false
}

// Update preferences given one voter's ranks
func recordPreferences(ranks []int) {
	for i := 0; i < candidateCount; i++ {
		for j := i + 1; j < candidateCount; j++ {
			preferences[ranks[i]][ranks[j]]++
		}
	}
}

// Record pairs of candidates where one is preferred over the other
func addPairs() {
	for i := 0; i < candidateCount; i++ {
		for j := i; j < candidateCount; j++ {
			if preferences[i][j] > preferences[j][i] {
				pairs[pairCount].winner = i
				pairs[pairCount].loser = j
				pairCount++
			} else if preferences[i][j] < preferences[j][i] {
				pairs[pairCount].winner = j
				pairs[pairCount].loser = i
				pairCount++
			}
		}
	}
}

// Sort pairs in decreasing order by strength of victory
func sortPairs() {
	// Go's equivalent of std::sort with a lambda function.
	// We slice the array `pairs[:pairCount]` to only sort the populated portion.
	sort.Slice(pairs[:pairCount], func(i, j int) bool {
		return preferences[pairs[i].winner][pairs[i].loser] > preferences[pairs[j].winner][pairs[j].loser]
	})
}

// Lock pairs into the candidate graph in order, without creating cycles
func lockPairs() {
	for i := 0; i < pairCount; i++ {
		if !makesCycle(pairs[i].winner, pairs[i].loser) {
			locked[pairs[i].winner][pairs[i].loser] = true
		}
	}
}

// Checks whether pairs of candidates makes a cycle
func makesCycle(start, end int) bool {
	if start == end {
		return true
	}
	
	for i := 0; i < candidateCount; i++ {
		if locked[end][i] {
			if makesCycle(start, i) {
				return true
			}
		}
	}
	
	return false
}

// Print the winner of the election
func printWinner() {
	// Initialize a slice with 'candidateCount' elements, all set to 0
	candidateEdges := make([]int, candidateCount)

	for i := 0; i < candidateCount; i++ {
		for j := 0; j < candidateCount; j++ {
			if locked[j][i] {
				candidateEdges[i]++
			}
		}
	}

	minEdges := candidateCount

	for i := 0; i < candidateCount; i++ {
		if candidateEdges[i] < minEdges {
			minEdges = candidateEdges[i]
		}
	}

	for i := 0; i < candidateCount; i++ {
		if candidateEdges[i] == minEdges {
			fmt.Println(candidates[i])
			break
		}
	}
}
