package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"unicode"
)

func main() {
	// Get text from user
	fmt.Print("Text: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	text := scanner.Text()

	// Count letters in text
	letterCount := countLetters(text)

	// Count words in text
	wordCount := countWords(text)

	// Count sentences in text
	sentenceCount := countSentences(text)

	// Compute Coleman-Liau index for text
	grade := computeIndex(letterCount, wordCount, sentenceCount)

	// Print grade level for text
	if grade < 1 {
		fmt.Println("Before Grade 1")
	} else if grade >= 16 {
		fmt.Println("Grade 16+")
	} else {
		fmt.Printf("Grade %d\n", grade)
	}
}

func countLetters(s string) int {
	count := 0
	// Range-based for loop makes character traversal much cleaner, yielding runes
	for _, c := range s {
		if unicode.IsLetter(c) {
			count++
		}
	}
	return count
}

func countWords(s string) int {
	if len(s) == 0 {
		return 0
	}

	// Since words are separated by spaces, there is always 
	// 1 more word than there are spaces.
	count := 1
	for _, c := range s {
		if c == ' ' {
			count++
		}
	}
	return count
}

func countSentences(s string) int {
	count := 0
	for _, c := range s {
		// Runes can be compared directly to character literals
		if c == '.' || c == '!' || c == '?' {
			count++
		}
	}
	return count
}

func computeIndex(letterC, wordC, sentenceC int) int {
	// Use explicit float64 conversion instead of C++ static_cast
	L := float64(letterC) / float64(wordC) * 100.0
	S := float64(sentenceC) / float64(wordC) * 100.0

	// Go defaults to float64 for decimal literals, so no 'f' suffix is needed
	index := 0.0588*L - 0.296*S - 15.8

	// math.Round returns a float64, so we cast it back to an int
	return int(math.Round(index))
}
