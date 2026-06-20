package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
)

func main() {
	// Checks whether key is valid, passing in command-line arguments
	if checkKey(os.Args) {
		// Get plaintext from user
		fmt.Print("plaintext:  ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		plaintext := scanner.Text()

		// Print ciphertext
		printCipherText(plaintext, os.Args[1])

		// Exit with success code
		os.Exit(0)
	} else {
		// Exit with error code
		os.Exit(1)
	}
}

// Checks if the provided command-line arguments contain a valid 26-letter key
func checkKey(args []string) bool {
	// args[0] is the program name, args[1] is the key
	if len(args) != 2 {
		fmt.Printf("Usage: %s key\n", args[0])
		return false
	}

	key := args[1]

	if len(key) != 26 {
		fmt.Println("Key must contain 26 characters.")
		return false
	}

	// In Go, we can use a map to check for duplicates much more efficiently 
	// than the nested loop used in the C++ version.
	seen := make(map[rune]bool)

	// Range over the string to inspect each character (rune)
	for _, char := range key {
		if !unicode.IsLetter(char) {
			fmt.Println("Key must contain only letters.")
			return false
		}

		lowerChar := unicode.ToLower(char)
		
		// If we have already seen this lowercase character, it's a duplicate
		if seen[lowerChar] {
			fmt.Println("Key must contain each letter exactly once.")
			return false
		}
		
		// Mark character as seen
		seen[lowerChar] = true
	}

	return true
}

// Maps the plaintext to the cipher and prints it
func printCipherText(text, cipher string) {
	// strings.Builder is Go's highly efficient way to build strings
	var result strings.Builder
	result.Grow(len(text)) // Optimizes memory allocation, matching result.reserve()

	// Range-based for loop: "For each rune 'c' in 'text'"
	for _, c := range text {
		if unicode.IsLetter(c) {
			if unicode.IsLower(c) {
				index := c - 'a' // Math on runes works just like chars in C++
				// Extract the corresponding cipher byte, cast to rune, and make lowercase
				cipherChar := unicode.ToLower(rune(cipher[index]))
				result.WriteRune(cipherChar)
			} else {
				index := c - 'A'
				// Extract the corresponding cipher byte, cast to rune, and make uppercase
				cipherChar := unicode.ToUpper(rune(cipher[index]))
				result.WriteRune(cipherChar)
			}
		} else {
			// Append non-alphabetic characters exactly as they are
			result.WriteRune(c)
		}
	}

	// Print the fully built string at once
	fmt.Printf("ciphertext: %s\n", result.String())
}
