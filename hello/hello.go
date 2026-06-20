package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	// Get name from user
	fmt.Print("What's your name? ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	name := scanner.Text()

	// Prints greeting to user
	fmt.Printf("hello, %s\n", name)
}
