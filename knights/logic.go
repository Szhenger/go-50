package main

import (
	"fmt"
	"strings"
	"unicode"
)

// ============================================================================
// Interface: Sentence
// ============================================================================

type Sentence interface {
	Evaluate(model map[string]bool) bool
	Formula() string
	Symbols() map[string]struct{}
	Repr() string
}

func Balanced(s string) bool {
	count := 0
	for _, c := range s {
		if c == '(' {
			count++
		} else if c == ')' {
			if count <= 0 {
				return false
			}
			count--
		}
	}
	return count == 0
}

func Parenthesize(s string) string {
	if s == "" {
		return s
	}
	isAlpha := true
	for _, r := range s {
		if !unicode.IsLetter(r) {
			isAlpha = false
			break
		}
	}
	if isAlpha {
		return s
	}
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		if Balanced(s[1 : len(s)-1]) {
			return s
		}
	}
	return "(" + s + ")"
}

// ============================================================================
// Logical Sentence Subclasses
// ============================================================================

type Symbol struct{ Name string }

func (s Symbol) Evaluate(m map[string]bool) bool { return m[s.Name] }
func (s Symbol) Formula() string                { return s.Name }
func (s Symbol) Symbols() map[string]struct{}   { return map[string]struct{}{s.Name: {}} }
func (s Symbol) Repr() string                   { return s.Name }

type Not struct{ Operand Sentence }

func (n Not) Evaluate(m map[string]bool) bool      { return !n.Operand.Evaluate(m) }
func (n Not) Formula() string                      { return "¬" + Parenthesize(n.Operand.Formula()) }
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
	if len(a.Conjuncts) == 0 { return "" }
	res := Parenthesize(a.Conjuncts[0].Formula())
	for i := 1; i < len(a.Conjuncts); i++ {
		res += " ∧ " + Parenthesize(a.Conjuncts[i].Formula())
	}
	return res
}
func (a And) Symbols() map[string]struct{} {
	syms := make(map[string]struct{})
	for _, c := range a.Conjuncts {
		for s := range c.Symbols() { syms[s] = struct{}{} }
	}
	return syms
}
func (a And) Repr() string {
	parts := []string{}
	for _, c := range a.Conjuncts { parts = append(parts, c.Repr()) }
	return "And(" + strings.Join(parts, ", ") + ")"
}

type Implication struct{ Antecedent, Consequent Sentence }

func (i Implication) Evaluate(m map[string]bool) bool {
	return !i.Antecedent.Evaluate(m) || i.Consequent.Evaluate(m)
}
func (i Implication) Formula() string {
	return Parenthesize(i.Antecedent.Formula()) + " => " + Parenthesize(i.Consequent.Formula())
}
func (i Implication) Symbols() map[string]struct{} {
	s := i.Antecedent.Symbols()
	for k := range i.Consequent.Symbols() { s[k] = struct{}{} }
	return s
}
func (i Implication) Repr() string {
	return "Implication(" + i.Antecedent.Repr() + ", " + i.Consequent.Repr() + ")"
}

// ============================================================================
// Model Checking
// ============================================================================

func checkAll(kb, query Sentence, symbols []string, model map[string]bool) bool {
	if len(symbols) == 0 {
		if kb.Evaluate(model) {
			return query.Evaluate(model)
		}
		return true
	}

	p := symbols[0]
	rest := symbols[1:]

	// True branch
	mTrue := make(map[string]bool)
	for k, v := range model { mTrue[k] = v }
	mTrue[p] = true

	// False branch
	mFalse := make(map[string]bool)
	for k, v := range model { mFalse[k] = v }
	mFalse[p] = false

	return checkAll(kb, query, rest, mTrue) && checkAll(kb, query, rest, mFalse)
}

func ModelCheck(kb, query Sentence) bool {
	symMap := kb.Symbols()
	for s := range query.Symbols() { symMap[s] = struct{}{} }

	symbols := []string{}
	for s := range symMap { symbols = append(symbols, s) }

	return checkAll(kb, query, symbols, make(map[string]bool))
}

func main() {
	A := Symbol{"A"}
	B := Symbol{"B"}
	KB := And{[]Sentence{Implication{A, B}, A}}

	fmt.Printf("Knowledge Base Formula: %s\n", KB.Formula())
	if ModelCheck(KB, B) {
		fmt.Println("Success: The Knowledge Base entails the query!")
	} else {
		fmt.Println("Failure: The Knowledge Base does NOT entail the query.")
	}
}
