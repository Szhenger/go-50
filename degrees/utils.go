package main

import (
	"errors"
)

// ============================================================================
// Search Node
// ============================================================================

// Node represents a generic search node.
// State must be comparable so we can check for equality in the frontier.
type Node[State comparable, Action any] struct {
	State  State
	Parent *Node[State, Action]
	
	// A pointer is used to represent std::optional. 
	// nil means the root/initial node has no action that led to it.
	Action *Action
}

// NewNode acts as the constructor.
func NewNode[State comparable, Action any](state State, parent *Node[State, Action], action *Action) *Node[State, Action] {
	return &Node[State, Action]{
		State:  state,
		Parent: parent,
		Action: action,
	}
}

// ============================================================================
// Stack Frontier (Depth-First Search)
// ============================================================================

// StackFrontier implements a Last-In, First-Out (LIFO) search frontier.
type StackFrontier[State comparable, Action any] struct {
	// A standard slice acts as our deque.
	frontier []*Node[State, Action]
}

func (s *StackFrontier[State, Action]) Add(node *Node[State, Action]) {
	s.frontier = append(s.frontier, node)
}

func (s *StackFrontier[State, Action]) ContainsState(state State) bool {
	for _, node := range s.frontier {
		if node.State == state {
			return true
		}
	}
	return false
}

func (s *StackFrontier[State, Action]) Empty() bool {
	return len(s.frontier) == 0
}

func (s *StackFrontier[State, Action]) Remove() (*Node[State, Action], error) {
	if s.Empty() {
		return nil, errors.New("empty frontier")
	}
	
	// Stack behavior: Last-In, First-Out (LIFO)
	lastIndex := len(s.frontier) - 1
	node := s.frontier[lastIndex]
	
	// Explicitly nil the array position to avoid memory leaks before slicing
	s.frontier[lastIndex] = nil 
	s.frontier = s.frontier[:lastIndex]
	
	return node, nil
}

// ============================================================================
// Queue Frontier (Breadth-First Search)
// ============================================================================

// QueueFrontier implements a First-In, First-Out (FIFO) search frontier.
// By embedding StackFrontier, it inherits Add, ContainsState, and Empty.
type QueueFrontier[State comparable, Action any] struct {
	StackFrontier[State, Action]
}

// Remove overrides the Stack behavior to process from the front.
func (q *QueueFrontier[State, Action]) Remove() (*Node[State, Action], error) {
	if q.Empty() {
		return nil, errors.New("empty frontier")
	}
	
	// Queue behavior: First-In, First-Out (FIFO)
	node := q.frontier[0]
	
	// Explicitly nil the array position to avoid memory leaks
	q.frontier[0] = nil
	q.frontier = q.frontier[1:]
	
	return node, nil
}
