package main

import (
	"fmt"
	"strings"
)

// Sentence defines the interface for all logical structures.
type Sentence interface {
	Evaluate(model map[string]bool) bool
	Formula() string
	Symbols() map[string]struct{}
	Repr() string
}

// ============================================================================
// Logic Subclasses
// ============================================================================

type Symbol struct{ Name string }

func (s Symbol) Evaluate(m map[string]bool) bool { return m[s.Name] }
func (s Symbol) Formula() string                { return s.Name }
func (s Symbol) Symbols() map[string]struct{}   { return map[string]struct{}{s.Name: {}} }
func (s Symbol) Repr() string                   { return s.Name }

type Not struct{ Operand Sentence }

func (n Not) Evaluate(m map[string]bool) bool      { return !n.Operand.Evaluate(m) }
func (n Not) Formula() string                      { return "¬(" + n.Operand.Formula() + ")" }
func (n Not) Symbols() map[string]struct{}         { return n.Operand.Symbols() }
func (n Not) Repr() string                         { return "Not(" + n.Operand.Repr() + ")" }

type And struct{ Conjuncts []Sentence }

func (a And) Evaluate(m map[string]bool) bool {
	for _, c := range a.Conjuncts {
		if !c.Evaluate(m) { return false }
	}
	return true
}
func (a And) Formula() string {
	parts := make([]string, len(a.Conjuncts))
	for i, c := range a.Conjuncts { parts[i] = "(" + c.Formula() + ")" }
	return strings.Join(parts, " ∧ ")
}
func (a And) Symbols() map[string]struct{} {
	s := make(map[string]struct{})
	for _, c := range a.Conjuncts {
		for sym := range c.Symbols() { s[sym] = struct{}{} }
	}
	return s
}
func (a And) Repr() string {
	parts := make([]string, len(a.Conjuncts))
	for i, c := range a.Conjuncts { parts[i] = c.Repr() }
	return "And(" + strings.Join(parts, ", ") + ")"
}

type Implication struct{ Ant, Cons Sentence }

func (i Implication) Evaluate(m map[string]bool) bool { return !i.Ant.Evaluate(m) || i.Cons.Evaluate(m) }
func (i Implication) Formula() string                { return "(" + i.Ant.Formula() + ") => (" + i.Cons.Formula() + ")" }
func (i Implication) Symbols() map[string]struct{} {
	s := i.Ant.Symbols()
	for k := range i.Cons.Symbols() { s[k] = struct{}{} }
	return s
}
func (i Implication) Repr() string { return "Implication(" + i.Ant.Repr() + ", " + i.Cons.Repr() + ")" }

// ============================================================================
// Logic Engine
// ============================================================================

func checkAll(kb, query Sentence, symbols []string, model map[string]bool) bool {
	if len(symbols) == 0 {
		if kb.Evaluate(model) { return query.Evaluate(model) }
		return true
	}

	p := symbols[0]
	rest := symbols[1:]

	// Branching: model P=true vs model P=false
	mTrue := copyModel(model); mTrue[p] = true
	mFalse := copyModel(model); mFalse[p] = false

	return checkAll(kb, query, rest, mTrue) && checkAll(kb, query, rest, mFalse)
}

func copyModel(m map[string]bool) map[string]bool {
	newM := make(map[string]bool)
	for k, v := range m { newM[k] = v }
	return newM
}

func ModelCheck(kb, query Sentence) bool {
	symSet := kb.Symbols()
	for s := range query.Symbols() { symSet[s] = struct{}{} }
	symbols := make([]string, 0, len(symSet))
	for s := range symSet { symbols = append(symbols, s) }
	return checkAll(kb, query, symbols, make(map[string]bool))
}

func main() {
	A := Symbol{"A"}
	B := Symbol{"B"}
	
	// KB: (A => B) ∧ A
	KB := And{[]Sentence{Implication{A, B}, A}}

	if ModelCheck(KB, B) {
		fmt.Println("Knowledge Base entails B")
	}
}
