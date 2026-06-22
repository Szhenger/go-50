package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Corpus maps a page name to a set of links (represented as a map[string]bool)
type Corpus map[string]map[string]bool

// ProbDist maps a page name to its corresponding probability rank
type ProbDist map[string]float64

// Constants
const (
	Damping = 0.85
	Samples = 10000
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: ./pagerank corpus")
		os.Exit(1)
	}

	directory := os.Args[1]
	corpus, err := crawl(directory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading corpus: %v\n", err)
		os.Exit(1)
	}

	// Get sorted keys for predictable output ordering
	var pages []string
	for page := range corpus {
		pages = append(pages, page)
	}
	sort.Strings(pages)

	sampledRanks := samplePagerank(corpus, Damping, Samples)
	fmt.Printf("PageRank Results from Sampling (n = %d)\n", Samples)
	for _, page := range pages {
		fmt.Printf("  %s: %.4f\n", page, sampledRanks[page])
	}

	iteratedRanks := iteratePagerank(corpus, Damping)
	fmt.Println("PageRank Results from Iteration")
	for _, page := range pages {
		fmt.Printf("  %s: %.4f\n", page, iteratedRanks[page])
	}
}

func crawl(directory string) (Corpus, error) {
	pages := make(Corpus)

	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	// Regex to extract links matching the C++ pattern
	linkRe := regexp.MustCompile(`<a\s+(?:[^>]*?)href="([^"]*)"`)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".html" {
			continue
		}

		filename := entry.Name()
		path := filepath.Join(directory, filename)
		
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		contents := string(contentBytes)

		pages[filename] = make(map[string]bool)
		matches := linkRe.FindAllStringSubmatch(contents, -1)
		
		for _, match := range matches {
			if len(match) > 1 {
				link := match[1]
				if link != filename {
					pages[filename][link] = true
				}
			}
		}
	}

	// Only include links to other pages present in the corpus
	for filename, links := range pages {
		for link := range links {
			if _, exists := pages[link]; !exists {
				delete(links, link)
			}
		}
	}

	return pages, nil
}

func transitionModel(corpus Corpus, page string, dampingFactor float64) ProbDist {
	probDist := make(ProbDist)
	
	for state := range corpus {
		probDist[state] = 0.0
	}

	linkedPages, exists := corpus[page]
	if exists {
		numPages := float64(len(corpus))
		numLinks := float64(len(linkedPages))

		if numLinks > 0 {
			for state := range linkedPages {
				probDist[state] += dampingFactor / numLinks
			}
			for state := range corpus {
				probDist[state] += (1.0 - dampingFactor) / numPages
			}
		} else {
			for state := range corpus {
				probDist[state] += 1.0 / numPages
			}
		}
	}
	
	return probDist
}

func samplePagerank(corpus Corpus, dampingFactor float64, n int) ProbDist {
	var pages []string
	samples := make(map[string]int)

	for page := range corpus {
		pages = append(pages, page)
		samples[page] = 0
	}
	sort.Strings(pages)

	// Randomly choose the initial sample page
	sample := pages[rand.Intn(len(pages))]
	samples[sample]++

	// Randomly choose remaining samples based on transition distribution weights
	for i := 1; i < n; i++ {
		model := transitionModel(corpus, sample, dampingFactor)
		
		var keys []string
		var weights []float64
		for key, weight := range model {
			keys = append(keys, key)
			weights = append(weights, weight)
		}

		sample = chooseWeighted(keys, weights)
		samples[sample]++
	}

	pageRanks := make(ProbDist)
	for _, page := range pages {
		pageRanks[page] = float64(samples[page]) / float64(n)
	}

	return pageRanks
}

// Mimics std::discrete_distribution by picking a key proportional to its weight
func chooseWeighted(keys []string, weights []float64) string {
	var totalWeight float64
	for _, w := range weights {
		totalWeight += w
	}

	r := rand.Float64() * totalWeight
	var cursor float64
	for i, w := range weights {
		cursor += w
		if r <= cursor {
			return keys[i]
		}
	}
	return keys[len(keys)-1]
}

func iteratePagerank(corpus Corpus, dampingFactor float64) ProbDist {
	pageRanks := make(ProbDist)
	numPages := float64(len(corpus))

	for page := range corpus {
		pageRanks[page] = 0.0
	}

	if numPages > 0 {
		for page := range corpus {
			pageRanks[page] = 1.0 / numPages
		}

		numLinks := make(map[string]float64)
		for page, links := range corpus {
			if len(links) == 0 {
				numLinks[page] = numPages
			} else {
				numLinks[page] = float64(len(links))
			}
		}

		// Sort keys to ensure exact matching sequence with C++ std::map updates
		var sortedPages []string
		for page := range corpus {
			sortedPages = append(sortedPages, page)
		}
		sort.Strings(sortedPages)

		iterate := true
		for iterate {
			iterate = false
			firstCondition := (1.0 - dampingFactor) / numPages

			for _, page := range sortedPages {
				currentRank := pageRanks[page]
				secondCondition := 0.0

				for _, linkingPage := range sortedPages {
					links := corpus[linkingPage]
					if links[page] || len(links) == 0 {
						secondCondition += pageRanks[linkingPage] / numLinks[linkingPage]
					}
				}
				
				secondCondition *= dampingFactor
				newRank := firstCondition + secondCondition
				pageRanks[page] = newRank

				if math.Abs(newRank-currentRank) > 0.001 {
					iterate = true
				}
			}
		}
	}

	return pageRanks
}
