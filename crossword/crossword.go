package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ============================================================================
// Core Crossword Structures
// ============================================================================

// Direction replaces the enum class
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

// Cell represents an (i, j) coordinate
type Cell struct {
	I, J int
}

// Variable represents a word slot in the crossword.
// The 'cells' field is omitted to ensure the struct remains comparable
// and usable as a map key.
type Variable struct {
	I, J   int
	Dir    Direction
	Length int
}

// Cells dynamically generates the coordinate cells for the variable.
func (v Variable) Cells() []Cell {
	cells := make([]Cell, 0, v.Length)
	for k := 0; k < v.Length; k++ {
		if v.Dir == Down {
			cells = append(cells, Cell{v.I + k, v.J})
		} else {
			cells = append(cells, Cell{v.I, v.J + k})
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

// String acts as to_string() and fulfills the fmt.Stringer interface
func (v Variable) String() string {
	return fmt.Sprintf("(%d, %d) %s : %d", v.I, v.J, v.Dir, v.Length)
}

// Repr acts as to_repr()
func (v Variable) Repr() string {
	return fmt.Sprintf("Variable(%d, %d, '%s', %d)", v.I, v.J, v.Dir, v.Length)
}

// VarPair is used as a composite key for the overlaps map
type VarPair struct {
	V1, V2 Variable
}

// Overlap stores the intersection indices
type Overlap struct {
	First, Second int
}

type Crossword struct {
	Height    int
	Width     int
	Structure [][]bool
	Words     map[string]bool
	Variables map[Variable]bool
	Overlaps  map[VarPair]*Overlap
}

func NewCrossword(structureFile, wordsFile string) (*Crossword, error) {
	// 1. Determine structure of crossword
	sFile, err := os.Open(structureFile)
	if err != nil {
		return nil, fmt.Errorf("could not open structure file: %w", err)
	}
	defer sFile.Close()

	var contents []string
	scanner := bufio.NewScanner(sFile)
	for scanner.Scan() {
		// Trim carriage returns to handle Windows CRLF files
		line := strings.TrimRight(scanner.Text(), "\r")
		contents = append(contents, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
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
	for i := 0; i < cw.Height; i++ {
		cw.Structure[i] = make([]bool, cw.Width)
		for j := 0; j < cw.Width; j++ {
			if j < len(contents[i]) && contents[i][j] == '_' {
				cw.Structure[i][j] = true
			}
		}
	}

	// 2. Save vocabulary list
	wFile, err := os.Open(wordsFile)
	if err != nil {
		return nil, fmt.Errorf("could not open words file: %w", err)
	}
	defer wFile.Close()

	wScanner := bufio.NewScanner(wFile)
	for wScanner.Scan() {
		line := strings.TrimRight(wScanner.Text(), "\r")
		if line != "" {
			cw.Words[strings.ToUpper(line)] = true
		}
	}
	if err := wScanner.Err(); err != nil {
		return nil, err
	}

	// 3. Determine variable set
	for i := 0; i < cw.Height; i++ {
		for j := 0; j < cw.Width; j++ {

			// Vertical words
			startsWordDown := cw.Structure[i][j] && (i == 0 || !cw.Structure[i-1][j])
			if startsWordDown {
				length := 1
				for k := i + 1; k < cw.Height; k++ {
					if cw.Structure[k][j] {
						length++
					} else {
						break
					}
				}
				if length > 1 {
					cw.Variables[Variable{i, j, Down, length}] = true
				}
			}

			// Horizontal words
			startsWordAcross := cw.Structure[i][j] && (j == 0 || !cw.Structure[i][j-1])
			if startsWordAcross {
				length := 1
				for k := j + 1; k < cw.Width; k++ {
					if cw.Structure[i][k] {
						length++
					} else {
						break
					}
				}
				if length > 1 {
					cw.Variables[Variable{i, j, Across, length}] = true
				}
			}
		}
	}

	// 4. Compute overlaps for each word
	for v1 := range cw.Variables {
		for v2 := range cw.Variables {
			if v1 == v2 {
				continue
			}

			var intersection *Overlap
			v1Cells := v1.Cells()
			v2Cells := v2.Cells()

			// Check cross-reference of coordinate pairs
			for idx1, c1 := range v1Cells {
				for idx2, c2 := range v2Cells {
					if c1 == c2 {
						intersection = &Overlap{First: idx1, Second: idx2}
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

// Neighbors returns a slice of variables that overlap with the given variable
func (c *Crossword) Neighbors(v Variable) []Variable {
	var result []Variable
	for other := range c.Variables {
		if v == other {
			continue
		}
		if c.Overlaps[VarPair{other, v}] != nil {
			result = append(result, other)
		}
	}
	return result
}
