package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// ============================================================================
// Core Data Structures
// ============================================================================

type Person struct {
	Name   string
	Birth  string
	Movies map[string]struct{}
}

type Movie struct {
	Title string
	Year  string
	Stars map[string]struct{}
}

type Node struct {
	State  string // Person ID
	Parent *Node
	Action string // Movie ID
}

type Pair struct {
	First  string // Movie ID
	Second string // Person ID
}

type DegreesOfSeparation struct {
	namesMap map[string]map[string]struct{}
	people   map[string]*Person
	movies   map[string]*Movie
}

func NewDegreesOfSeparation() *DegreesOfSeparation {
	return &DegreesOfSeparation{
		namesMap: make(map[string]map[string]struct{}),
		people:   make(map[string]*Person),
		movies:   make(map[string]*Movie),
	}
}

// ============================================================================
// Loading & Graph Logic
// ============================================================================

func (ds *DegreesOfSeparation) LoadData(directory string) error {
	// Helper method to read CSV and map columns dynamically
	readCSV := func(filename string, rowHandler func(getCol func(string) string)) error {
		path := filepath.Join(directory, filename)
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("could not open %s: %w", path, err)
		}
		defer file.Close()

		reader := csv.NewReader(file)
		// Read header to dynamically map columns
		headers, err := reader.Read()
		if err != nil {
			return err
		}

		headerMap := make(map[string]int)
		for i, h := range headers {
			headerMap[h] = i
		}

		for {
			row, err := reader.Read()
			if err != nil {
				break // EOF or error
			}

			getCol := func(colName string) string {
				if idx, exists := headerMap[colName]; exists && idx < len(row) {
					return row[idx]
				}
				return ""
			}
			rowHandler(getCol)
		}
		return nil
	}

	// 1. Load people
	err := readCSV("people.csv", func(getCol func(string) string) {
		id := getCol("id")
		name := getCol("name")
		birth := getCol("birth")

		ds.people[id] = &Person{
			Name:   name,
			Birth:  birth,
			Movies: make(map[string]struct{}),
		}

		lowerName := strings.ToLower(name)
		if ds.namesMap[lowerName] == nil {
			ds.namesMap[lowerName] = make(map[string]struct{})
		}
		ds.namesMap[lowerName][id] = struct{}{}
	})
	if err != nil {
		return err
	}

	// 2. Load movies
	err = readCSV("movies.csv", func(getCol func(string) string) {
		id := getCol("id")
		ds.movies[id] = &Movie{
			Title: getCol("title"),
			Year:  getCol("year"),
			Stars: make(map[string]struct{}),
		}
	})
	if err != nil {
		return err
	}

	// 3. Load stars
	err = readCSV("stars.csv", func(getCol func(string) string) {
		personID := getCol("person_id")
		movieID := getCol("movie_id")

		if p, exists := ds.people[personID]; exists {
			p.Movies[movieID] = struct{}{}
		}
		if m, exists := ds.movies[movieID]; exists {
			m.Stars[personID] = struct{}{}
		}
	})
	if err != nil {
		return err
	}

	return nil
}

func (ds *DegreesOfSeparation) neighborsForPerson(personID string) []Pair {
	var neighbors []Pair
	p, exists := ds.people[personID]
	if !exists {
		return neighbors
	}

	for movieID := range p.Movies {
		if m, exists := ds.movies[movieID]; exists {
			for starID := range m.Stars {
				neighbors = append(neighbors, Pair{First: movieID, Second: starID})
			}
		}
	}
	return neighbors
}

func (ds *DegreesOfSeparation) PersonIDForName(name string) (string, bool) {
	lowerName := strings.ToLower(name)
	personIDs, exists := ds.namesMap[lowerName]
	
	if !exists || len(personIDs) == 0 {
		return "", false
	}

	if len(personIDs) == 1 {
		for id := range personIDs {
			return id, true
		}
	}

	// Handle ambiguity
	fmt.Printf("Which '%s'?\n", name)
	for id := range personIDs {
		p := ds.people[id]
		fmt.Printf("ID: %s, Name: %s, Birth: %s\n", id, p.Name, p.Birth)
	}

	fmt.Print("Intended Person ID: ")
	reader := bufio.NewReader(os.Stdin)
	intendedID, _ := reader.ReadString('\n')
	intendedID = strings.TrimSpace(intendedID)

	if _, ok := personIDs[intendedID]; ok {
		return intendedID, true
	}

	return "", false
}

func (ds *DegreesOfSeparation) ShortestPath(source, target string) ([]Pair, bool) {
	if source == target {
		return []Pair{}, true
	}

	var frontier []*Node
	explored := make(map[string]struct{})
	
	// Frontier states optimization for O(1) containment checks
	frontierStates := make(map[string]struct{})

	for _, neighbor := range ds.neighborsForPerson(source) {
		frontier = append(frontier, &Node{State: neighbor.Second, Parent: nil, Action: neighbor.First})
		frontierStates[neighbor.Second] = struct{}{}
	}

	for len(frontier) > 0 {
		// Pop front
		node := frontier[0]
		frontier = frontier[1:]
		
		delete(frontierStates, node.State)
		explored[node.State] = struct{}{}

		if node.State == target {
			var path []Pair
			curr := node
			for curr != nil {
				path = append(path, Pair{First: curr.Action, Second: curr.State})
				curr = curr.Parent
			}
			slices.Reverse(path)
			return path, true
		}

		for _, neighbor := range ds.neighborsForPerson(node.State) {
			_, isExplored := explored[neighbor.Second]
			_, isFrontier := frontierStates[neighbor.Second]

			if !isExplored && !isFrontier {
				frontier = append(frontier, &Node{State: neighbor.Second, Parent: node, Action: neighbor.First})
				frontierStates[neighbor.Second] = struct{}{}
			}
		}
	}

	return nil, false
}

func (ds *DegreesOfSeparation) GetPerson(id string) *Person {
	return ds.people[id]
}

func (ds *DegreesOfSeparation) GetMovie(id string) *Movie {
	return ds.movies[id]
}

// ============================================================================
// Main Execution
// ============================================================================

func main() {
	if len(os.Args) > 2 {
		fmt.Fprintln(os.Stderr, "Usage: ./degrees [directory]")
		os.Exit(1)
	}

	directory := "large"
	if len(os.Args) == 2 {
		directory = os.Args[1]
	}

	ds := NewDegreesOfSeparation()

	fmt.Println("Loading data...")
	if err := ds.LoadData(directory); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Data loaded.")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Name: ")
	sourceName, _ := reader.ReadString('\n')
	source, ok := ds.PersonIDForName(strings.TrimSpace(sourceName))
	if !ok {
		fmt.Fprintln(os.Stderr, "Person not found.")
		os.Exit(1)
	}

	fmt.Print("Name: ")
	targetName, _ := reader.ReadString('\n')
	target, ok := ds.PersonIDForName(strings.TrimSpace(targetName))
	if !ok {
		fmt.Fprintln(os.Stderr, "Person not found.")
		os.Exit(1)
	}

	path, found := ds.ShortestPath(source, target)

	if !found {
		fmt.Println("Not connected.")
	} else {
		degrees := len(path)
		fmt.Printf("%d degrees of separation.\n", degrees)

		// Prepend source to match the printing loop logic
		fullPath := append([]Pair{{First: "", Second: source}}, path...)

		for i := 0; i < degrees; i++ {
			person1 := ds.GetPerson(fullPath[i].Second).Name
			person2 := ds.GetPerson(fullPath[i+1].Second).Name
			movie := ds.GetMovie(fullPath[i+1].First).Title

			fmt.Printf("%d: %s and %s starred in %s\n", i+1, person1, person2, movie)
		}
	}
}
