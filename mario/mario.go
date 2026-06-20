package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	// Get height of the desired pyramid between 1 and 8
	height := getHeight()

	// Prints the desired pyramid
	makePyramid(height)
}

// Returns an integer between 1 and 8
func getHeight() int {
	scanner := bufio.NewScanner(os.Stdin)
	var i int

	// Go's equivalent of a do-while loop
	for {
		fmt.Print("Height: ")

		// Read the entire line of input
		scanner.Scan()
		input := scanner.Text()

		// Attempt to parse the input as an integer (stripping surrounding whitespace)
		parsedInt, err := strconv.Atoi(strings.TrimSpace(input))

		// If it failed (e.g., user typed "apple"), err is non-nil.
		// The scanner already consumed the line, so garbage is inherently discarded.
		if err != nil {
			continue
		}

		i = parsedInt

		// Check bounds to break out of the loop
		if i >= 1 && i <= 8 {
			break
		}
	}

	return i
}

// Prints the desired pyramid of height n
func makePyramid(n int) {
	for i := 0; i < n; i++ {
		for j := 0; j < n+(i+3); j++ {
			if j < n-(i+1) || j > n+(i+2) || j == n || j == n+1 {
				fmt.Print(" ")
			} else {
				fmt.Print("#")
			}
		}
		fmt.Println() // Go's equivalent of std::println()
	}
}
