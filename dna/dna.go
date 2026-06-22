// Package main implements a program that identifies to whom a sequence of DNA belongs
package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

// A Row maps column names to either a string or an int.
// In Go, 'any' is an alias for 'interface{}', replacing C++'s std::variant.
type Row map[string]any

// longestMatch returns the length of the longest consecutive run of `subsequence`
// found inside `sequence`.
func longestMatch(sequence, subsequence string) int {
	longestRun := 0
	subseqLen := len(subsequence)
	seqLen := len(sequence)

	for i := 0; i < seqLen; i++ {
		count := 0
		for {
			start := i + count*subseqLen
			end := start + subseqLen

			if end <= seqLen && sequence[start:end] == subsequence {
				count++
			} else {
				break
			}
		}
		if count > longestRun {
			longestRun = count
		}
	}
	return longestRun
}

// parseCSV reads a CSV file and returns a slice of rows. Each row maps header
// names to 'any' (int when the cell is purely numeric, string otherwise).
func parseCSV(filename string) ([]Row, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("cannot open file '%s'", filename)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return []Row{}, nil
	}

	// ---- header row --------------------------------------------------------
	headers := records[0]
	var database []Row

	// ---- data rows ---------------------------------------------------------
	for _, record := range records[1:] {
		row := make(Row)
		for col, field := range record {
			if col >= len(headers) {
				break
			}

			// Mirror Python's try/except int() conversion.
			if value, err := strconv.Atoi(field); err == nil {
				row[headers[col]] = value
			} else {
				row[headers[col]] = field
			}
		}
		database = append(database, row)
	}

	return database, nil
}

func main() {
	// Check for command-line usage
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Usage: dna data.csv sequence.txt")
		os.Exit(1)
	}

	// Read database file
	database, err := parseCSV(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if len(database) == 0 {
		fmt.Fprintln(os.Stderr, "Error: database is empty.")
		os.Exit(1)
	}

	// Read DNA sequence file into a string
	seqBytes, err := os.ReadFile(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot open file '%s'\n", os.Args[2])
		os.Exit(1)
	}
	sequence := string(seqBytes)

	// Find longest match of each STR in DNA sequence.
	// STR columns are those whose first-row value is an int.
	sequenceData := make(map[string]int)
	for header, value := range database[0] {
		// Type assertion: checks if the underlying value is an int
		if _, ok := value.(int); ok {
			sequenceData[header] = longestMatch(sequence, header)
		}
	}

	// Check database for matching profiles
	for _, profile := range database {
		match := true

		// Replicates std::ranges::all_of behavior
		for header, expected := range sequenceData {
			val, exists := profile[header]
			if !exists {
				match = false
				break
			}

			// Extract integer and compare
			actual, isInt := val.(int)
			if !isInt || actual != expected {
				match = false
				break
			}
		}

		if match {
			if name, ok := profile["name"].(string); ok {
				fmt.Println(name)
				os.Exit(0)
			}
		}
	}

	fmt.Println("No match")
	os.Exit(0)
}
