package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
)

// ============================================================================
// Core Constants and Structures
// ============================================================================

const ProbsMutation = 0.01

var ProbsGene = map[int]float64{
	2: 0.01,
	1: 0.03,
	0: 0.96,
}

var ProbsTrait = map[int]map[bool]float64{
	2: {true: 0.65, false: 0.35},
	1: {true: 0.56, false: 0.44},
	0: {true: 0.01, false: 0.99},
}

type Person struct {
	Name   string
	Mother *string
	Father *string
	Trait  *bool
}

type Probabilities struct {
	Gene  map[int]float64
	Trait map[bool]float64
}

func NewProbabilities() *Probabilities {
	return &Probabilities{
		Gene:  map[int]float64{0: 0.0, 1: 0.0, 2: 0.0},
		Trait: map[bool]float64{true: 0.0, false: 0.0},
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func loadData(filename string) (map[string]Person, error) {
	data := make(map[string]Person)
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Skip header
	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	for {
		row, err := reader.Read()
		if err != nil {
			break // EOF or error
		}

		p := Person{Name: row[0]}

		if row[1] != "" {
			mother := row[1]
			p.Mother = &mother
		}
		if row[2] != "" {
			father := row[2]
			p.Father = &father
		}
		if row[3] == "1" {
			t := true
			p.Trait = &t
		} else if row[3] == "0" {
			t := false
			p.Trait = &t
		}

		data[p.Name] = p
	}

	return data, nil
}

// powerset generates all subsets of a given set (2^n combinations)
func powerset(s map[string]struct{}) []map[string]struct{} {
	var vec []string
	for k := range s {
		vec = append(vec, k)
	}

	n := len(vec)
	var result []map[string]struct{}

	for i := 0; i < (1 << n); i++ {
		subset := make(map[string]struct{})
		for j := 0; j < n; j++ {
			if (i & (1 << j)) != 0 {
				subset[vec[j]] = struct{}{}
			}
		}
		result = append(result, subset)
	}

	return result
}

// setDifference returns elements in 'a' that are not in 'b'
func setDifference(a, b map[string]struct{}) map[string]struct{} {
	result := make(map[string]struct{})
	for val := range a {
		if _, exists := b[val]; !exists {
			result[val] = struct{}{}
		}
	}
	return result
}

// ============================================================================
// Core Logic
// ============================================================================

func jointProbability(
	people map[string]Person,
	oneGene map[string]struct{},
	twoGenes map[string]struct{},
	haveTrait map[string]struct{},
) float64 {

	probability := 1.0

	for person, info := range people {
		mother := info.Mother
		father := info.Father

		motherProbability := ProbsMutation
		if mother != nil {
			if _, ok := oneGene[*mother]; ok {
				motherProbability = 0.5
			} else if _, ok := twoGenes[*mother]; ok {
				motherProbability = 1.0 - ProbsMutation
			}
		}

		fatherProbability := ProbsMutation
		if father != nil {
			if _, ok := oneGene[*father]; ok {
				fatherProbability = 0.5
			} else if _, ok := twoGenes[*father]; ok {
				fatherProbability = 1.0 - ProbsMutation
			}
		}

		_, hasOneGene := oneGene[person]
		_, hasTwoGenes := twoGenes[person]
		_, hasTheTrait := haveTrait[person]

		if hasOneGene {
			oneProbability := 1.0
			if mother == nil && father == nil {
				oneProbability *= ProbsGene[1]
			} else {
				oneProbability *= (motherProbability*(1.0-fatherProbability) + fatherProbability*(1.0-motherProbability))
			}

			if hasTheTrait {
				oneProbability *= ProbsTrait[1][true]
			} else {
				oneProbability *= ProbsTrait[1][false]
			}

			probability *= oneProbability

		} else if hasTwoGenes {
			twoProbability := 1.0
			if mother == nil && father == nil {
				twoProbability *= ProbsGene[2]
			} else {
				twoProbability *= motherProbability * fatherProbability
			}

			if hasTheTrait {
				twoProbability *= ProbsTrait[2][true]
			} else {
				twoProbability *= ProbsTrait[2][false]
			}

			probability *= twoProbability

		} else {
			zeroProbability := 1.0
			if mother == nil && father == nil {
				zeroProbability *= ProbsGene[0]
			} else {
				zeroProbability *= (1.0 - motherProbability) * (1.0 - fatherProbability)
			}

			if hasTheTrait {
				zeroProbability *= ProbsTrait[0][true]
			} else {
				zeroProbability *= ProbsTrait[0][false]
			}

			probability *= zeroProbability
		}
	}

	return probability
}

func update(
	probabilities map[string]*Probabilities,
	oneGene map[string]struct{},
	twoGenes map[string]struct{},
	haveTrait map[string]struct{},
	p float64,
) {
	for person, probs := range probabilities {
		gene := 0
		if _, ok := oneGene[person]; ok {
			gene = 1
		} else if _, ok := twoGenes[person]; ok {
			gene = 2
		}

		probs.Gene[gene] += p

		_, trait := haveTrait[person]
		probs.Trait[trait] += p
	}
}

func normalize(probabilities map[string]*Probabilities) {
	for _, probs := range probabilities {
		// Normalize genes
		geneSum := 0.0
		for i := 0; i < 3; i++ {
			geneSum += probs.Gene[i]
		}
		for i := 0; i < 3; i++ {
			probs.Gene[i] /= geneSum
		}

		// Normalize traits
		traitSum := probs.Trait[true] + probs.Trait[false]
		probs.Trait[true] /= traitSum
		probs.Trait[false] /= traitSum
	}
}

// ============================================================================
// Main Execution
// ============================================================================

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: ./heredity data.csv")
		os.Exit(1)
	}

	people, err := loadData(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading data: %v\n", err)
		os.Exit(1)
	}

	probabilities := make(map[string]*Probabilities)
	names := make(map[string]struct{})

	for person := range people {
		probabilities[person] = NewProbabilities()
		names[person] = struct{}{}
	}

	// Loop over all sets of people who might have the trait
	for _, haveTrait := range powerset(names) {

		failsEvidence := false
		for person := range names {
			_, traitInSet := haveTrait[person]
			
			if people[person].Trait != nil && *people[person].Trait != traitInSet {
				failsEvidence = true
				break
			}
		}
		if failsEvidence {
			continue
		}

		// Loop over all sets of people who might have the gene
		for _, oneGene := range powerset(names) {
			remainingNames := setDifference(names, oneGene)

			for _, twoGenes := range powerset(remainingNames) {
				p := jointProbability(people, oneGene, twoGenes, haveTrait)
				update(probabilities, oneGene, twoGenes, haveTrait, p)
			}
		}
	}

	normalize(probabilities)

	// Sort names alphabetically for deterministic output mapping
	var sortedNames []string
	for name := range names {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	// Print results mapping identically to the expected outputs
	for _, person := range sortedNames {
		fmt.Printf("%s:\n", person)
		fmt.Printf("  Gene:\n")
		for i := 2; i >= 0; i-- {
			fmt.Printf("    %d: %.4f\n", i, probabilities[person].Gene[i])
		}
		fmt.Printf("  Trait:\n")
		fmt.Printf("    True: %.4f\n", probabilities[person].Trait[true])
		fmt.Printf("    False: %.4f\n", probabilities[person].Trait[false])
	}
}
