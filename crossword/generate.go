package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

// ============================================================================
// Core Crossword Structures
// ============================================================================

type Direction int

const (
	Across Direction = iota
	Down
)

func (d Direction) String() string {
	if d == Across {
		return "across"
	}
	return "down"
}

type Cell struct {
	I, J int
}

type Variable struct {
	I, J   int
	Dir    Direction
	Length int
}

func (v Variable) Cells() []Cell {
	cells := make([]Cell, v.Length)
	for k := 0; k < v.Length; k++ {
		if v.Dir == Down {
			cells[k] = Cell{v.I + k, v.J}
		} else {
			cells[k] = Cell{v.I, v.J + k}
		}
	}
	return cells
}

// Compare enables deterministic sorting matching C++'s operator<=>
func (v Variable) Compare(other Variable) int {
	if v.I != other.I {
		return v.I - other.I
	}
	if v.J != other.J {
		return v.J - other.J
	}
	if v.Dir != other.Dir {
		return int(v.Dir) - int(other.Dir)
	}
	return v.Length - other.Length
}

type VarPair struct {
	V1, V2 Variable
}

type Overlap struct {
	First, Second int
}

type Crossword struct {
	Height, Width int
	Structure     [][]bool
	Words         map[string]bool
	Variables     map[Variable]bool
	Overlaps      map[VarPair]*Overlap
}

func NewCrossword(structureFile, wordsFile string) (*Crossword, error) {
	sFile, err := os.Open(structureFile)
	if err != nil {
		return nil, fmt.Errorf("could not open structure file: %w", err)
	}
	defer sFile.Close()

	var contents []string
	scanner := bufio.NewScanner(sFile)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		contents = append(contents, line)
	}

	cw := &Crossword{
		Height:    len(contents),
		Words:     make(map[string]bool),
		Variables: make(map[Variable]bool),
		Overlaps:  make(map[VarPair]*Overlap),
	}

	for _, l := range contents {
		if len(l) > cw.Width {
			cw.Width = len(l)
		}
	}

	cw.Structure = make([][]bool, cw.Height)
	for i := range cw.Structure {
		cw.Structure[i] = make([]bool, cw.Width)
		for j := 0; j < cw.Width; j++ {
			if j < len(contents[i]) && contents[i][j] == '_' {
				cw.Structure[i][j] = true
			}
		}
	}

	wFile, err := os.Open(wordsFile)
	if err != nil {
		return nil, fmt.Errorf("could not open words file: %w", err)
	}
	defer wFile.Close()

	scanner = bufio.NewScanner(wFile)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		if line != "" {
			cw.Words[strings.ToUpper(line)] = true
		}
	}

	// Find Variables (Contiguous spaces)
	for i := 0; i < cw.Height; i++ {
		for j := 0; j < cw.Width; j++ {
			if cw.Structure[i][j] && (i == 0 || !cw.Structure[i-1][j]) {
				length := 1
				for k := i + 1; k < cw.Height && cw.Structure[k][j]; k++ {
					length++
				}
				if length > 1 {
					cw.Variables[Variable{i, j, Down, length}] = true
				}
			}
			if cw.Structure[i][j] && (j == 0 || !cw.Structure[i][j-1]) {
				length := 1
				for k := j + 1; k < cw.Width && cw.Structure[i][k]; k++ {
					length++
				}
				if length > 1 {
					cw.Variables[Variable{i, j, Across, length}] = true
				}
			}
		}
	}

	// Pre-calculate Overlaps
	sortedVars := cw.SortedVariables()
	for _, v1 := range sortedVars {
		for _, v2 := range sortedVars {
			if v1 == v2 {
				continue
			}
			
			var intersection *Overlap
			v1Cells := v1.Cells()
			v2Cells := v2.Cells()
			
			for idx1, c1 := range v1Cells {
				for idx2, c2 := range v2Cells {
					if c1 == c2 {
						intersection = &Overlap{idx1, idx2}
						break
					}
				}
				if intersection != nil {
					break
				}
			}
			cw.Overlaps[VarPair{v1, v2}] = intersection
		}
	}

	return cw, nil
}

func (c *Crossword) Neighbors(v Variable) []Variable {
	var neighbors []Variable
	for other := range c.Variables {
		if v == other {
			continue
		}
		if c.Overlaps[VarPair{other, v}] != nil {
			neighbors = append(neighbors, other)
		}
	}
	return neighbors
}

// SortedVariables mimics std::set deterministic iteration
func (c *Crossword) SortedVariables() []Variable {
	vars := make([]Variable, 0, len(c.Variables))
	for v := range c.Variables {
		vars = append(vars, v)
	}
	sort.Slice(vars, func(i, j int) bool {
		return vars[i].Compare(vars[j]) < 0
	})
	return vars
}

// ============================================================================
// Crossword Creator & Solver Logic
// ============================================================================

type CrosswordCreator struct {
	Crossword *Crossword
	Domains   map[Variable]map[string]bool
}

func NewCrosswordCreator(cw *Crossword) *CrosswordCreator {
	cc := &CrosswordCreator{
		Crossword: cw,
		Domains:   make(map[Variable]map[string]bool),
	}
	for v := range cw.Variables {
		cc.Domains[v] = make(map[string]bool)
		for word := range cw.Words {
			cc.Domains[v][word] = true
		}
	}
	return cc
}

func (cc *CrosswordCreator) LetterGrid(assignment map[Variable]string) map[Cell]rune {
	letters := make(map[Cell]rune)
	for variable, word := range assignment {
		for k, char := range word {
			i := variable.I
			j := variable.J
			if variable.Dir == Down {
				i += k
			} else {
				j += k
			}
			letters[Cell{i, j}] = char
		}
	}
	return letters
}

func (cc *CrosswordCreator) Print(assignment map[Variable]string) {
	letters := cc.LetterGrid(assignment)
	for i := 0; i < cc.Crossword.Height; i++ {
		for j := 0; j < cc.Crossword.Width; j++ {
			if cc.Crossword.Structure[i][j] {
				if char, exists := letters[Cell{i, j}]; exists {
					fmt.Printf("%c", char)
				} else {
					fmt.Print(" ")
				}
			} else {
				fmt.Print("█")
			}
		}
		fmt.Println()
	}
}

func (cc *CrosswordCreator) Save(assignment map[Variable]string, filename string) {
	fmt.Printf("[Info] Image saving triggered for '%s'.\n", filename)
	// Equivalent to C++ note: True image writing logic omitted
}

func (cc *CrosswordCreator) EnforceNodeConsistency() {
	for v, wordSet := range cc.Domains {
		for word := range wordSet {
			if len(word) != v.Length {
				delete(cc.Domains[v], word)
			}
		}
	}
}

func (cc *CrosswordCreator) Revise(x, y Variable) bool {
	revised := false
	overlap := cc.Crossword.Overlaps[VarPair{x, y}]

	if overlap != nil {
		yChars := make(map[byte]bool)
		for yWord := range cc.Domains[y] {
			yChars[yWord[overlap.Second]] = true
		}

		for xWord := range cc.Domains[x] {
			if !yChars[xWord[overlap.First]] {
				delete(cc.Domains[x], xWord)
				revised = true
			}
		}
	}
	return revised
}

func (cc *CrosswordCreator) AC3() bool {
	var queue []VarPair
	sortedVars := cc.Crossword.SortedVariables()
	
	// Populate initial queue
	for _, v1 := range sortedVars {
		for _, v2 := range sortedVars {
			if v1 != v2 {
				queue = append(queue, VarPair{v1, v2})
			}
		}
	}

	for len(queue) > 0 {
		arc := queue[0]
		queue = queue[1:]

		if cc.Revise(arc.V1, arc.V2) {
			if len(cc.Domains[arc.V1]) == 0 {
				return false
			}
			
			// Enqueue neighbors (deterministic)
			neighbors := cc.Crossword.Neighbors(arc.V1)
			sort.Slice(neighbors, func(i, j int) bool {
				return neighbors[i].Compare(neighbors[j]) < 0
			})
			
			for _, v3 := range neighbors {
				if v3 != arc.V1 && v3 != arc.V2 {
					queue = append(queue, VarPair{v3, arc.V1})
				}
			}
		}
	}
	return true
}

func (cc *CrosswordCreator) AssignmentComplete(assignment map[Variable]string) bool {
	return len(assignment) == len(cc.Crossword.Variables)
}

func (cc *CrosswordCreator) Consistent(assignment map[Variable]string) bool {
	for v1, word1 := range assignment {
		if len(word1) != v1.Length {
			return false
		}
		for _, v2 := range cc.Crossword.Neighbors(v1) {
			overlap := cc.Crossword.Overlaps[VarPair{v1, v2}]
			if overlap != nil {
				if word2, assigned := assignment[v2]; assigned {
					if word1[overlap.First] != word2[overlap.Second] {
						return false
					}
				}
			}
		}
	}
	return true
}

func (cc *CrosswordCreator) OrderDomainValues(v Variable, assignment map[Variable]string) []string {
	var domain []string
	for word := range cc.Domains[v] {
		domain = append(domain, word)
	}

	eliminations := make(map[string]int)
	neighbors := cc.Crossword.Neighbors(v)

	for _, word := range domain {
		for _, neighbor := range neighbors {
			if _, exists := assignment[neighbor]; exists {
				continue
			}

			overlap := cc.Crossword.Overlaps[VarPair{v, neighbor}]
			if overlap != nil {
				for neighborWord := range cc.Domains[neighbor] {
					if word[overlap.First] != neighborWord[overlap.Second] {
						eliminations[word]++
					}
				}
			}
		}
	}

	// Sort matching C++ stable priority list
	sort.SliceStable(domain, func(i, j int) bool {
		if eliminations[domain[i]] != eliminations[domain[j]] {
			return eliminations[domain[i]] < eliminations[domain[j]]
		}
		// Tie-breaker fallback
		return domain[i] < domain[j]
	})

	return domain
}

func (cc *CrosswordCreator) SelectUnassignedVariable(assignment map[Variable]string) Variable {
	var unassigned []Variable
	for v := range cc.Domains {
		if _, exists := assignment[v]; !exists {
			unassigned = append(unassigned, v)
		}
	}

	// Sort unassigned (MRV heuristic followed by Degree heuristic)
	sort.Slice(unassigned, func(i, j int) bool {
		a, b := unassigned[i], unassigned[j]
		sizeA, sizeB := len(cc.Domains[a]), len(cc.Domains[b])
		
		if sizeA != sizeB {
			return sizeA < sizeB
		}
		
		neighborsA := len(cc.Crossword.Neighbors(a))
		neighborsB := len(cc.Crossword.Neighbors(b))
		if neighborsA != neighborsB {
			return neighborsA < neighborsB
		}
		
		// Fallback for strict cross-run determinism
		return a.Compare(b) < 0
	})

	return unassigned[0]
}

func (cc *CrosswordCreator) Backtrack(assignment map[Variable]string) map[Variable]string {
	if cc.AssignmentComplete(assignment) {
		return assignment
	}

	v := cc.SelectUnassignedVariable(assignment)
	for _, value := range cc.OrderDomainValues(v, assignment) {
		assignment[v] = value
		
		if cc.Consistent(assignment) {
			result := cc.Backtrack(assignment)
			if result != nil {
				return result
			}
		}
		
		// Backtrack (Delete key mirrors std::map erase)
		delete(assignment, v)
	}
	
	return nil
}

func (cc *CrosswordCreator) Solve() map[Variable]string {
	cc.EnforceNodeConsistency()
	cc.AC3()
	return cc.Backtrack(make(map[Variable]string))
}

// ============================================================================
// Execution Core
// ============================================================================

func main() {
	if len(os.Args) != 3 && len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "Usage: ./generate structure words [output]")
		os.Exit(1)
	}

	structureFile := os.Args[1]
	wordsFile := os.Args[2]
	outputFile := ""
	if len(os.Args) == 4 {
		outputFile = os.Args[3]
	}

	crossword, err := NewCrossword(structureFile, wordsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	creator := NewCrosswordCreator(crossword)
	assignment := creator.Solve()

	if assignment == nil {
		fmt.Println("No solution.")
	} else {
		creator.Print(assignment)
		if outputFile != "" {
			creator.Save(assignment, outputFile)
		}
	}
}
