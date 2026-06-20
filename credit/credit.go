package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	// Get number
	number := getLong("Number: ")

	// Initialize checksum, length, and startingdigits
	checksum := 0
	length := 0
	var startingdigits int64 = 0
	var digit int64 = 0

	// Compute checksum, length, and startingdigits of number
	for number > 0 {
		if startingdigits == 0 && number < 100 {
			startingdigits = number
		}
		digit = number % 10
		number = number / 10
		length++

		if length%2 == 0 {
			if 2*digit > 9 {
				// Go requires explicit conversion when adding an int64 to an int
				checksum += int((2*digit)/10 + (2*digit)%10)
			} else {
				checksum += int(2 * digit)
			}
		} else {
			checksum += int(digit)
		}
	}

	// Check number and bank
	if checksum%10 != 0 {
		fmt.Println("INVALID")
	} else if length == 13 || length == 16 {
		if length == 16 && startingdigits > 50 && startingdigits < 56 {
			fmt.Println("MASTERCARD")
		} else if startingdigits/10 == 4 {
			fmt.Println("VISA")
		} else {
			fmt.Println("INVALID")
		}
	} else if length == 15 {
		if startingdigits == 34 || startingdigits == 37 {
			fmt.Println("AMEX")
		} else {
			fmt.Println("INVALID")
		}
	} else {
		fmt.Println("INVALID")
	}
}

// Safely gets a 64-bit integer from the user, rejecting bad input
func getLong(prompt string) int64 {
	scanner := bufio.NewScanner(os.Stdin)
	var value int64

	for {
		// Print the prompt
		fmt.Print(prompt)

		// Attempt to read the entire line
		scanner.Scan()
		input := scanner.Text()

		// Attempt to parse the input as a base-10, 64-bit integer
		parsedInt, err := strconv.ParseInt(strings.TrimSpace(input), 10, 64)

		// If successful, break out of the infinite loop
		if err == nil {
			value = parsedInt
			break
		}
		// If err != nil, the loop repeats, naturally discarding the garbage input
	}

	return value
}
