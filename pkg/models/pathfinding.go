// pathfinding.go
// Create this file in the pkg/models directory
package models

import (
	"container/heap"
	"math"
)

type Node struct {
	Position Position
	f, g, h  float64
	parent   *Node
	index    int
}

type PriorityQueue []*Node

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].f < pq[j].f
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	node := x.(*Node)
	node.index = n
	*pq = append(*pq, node)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	node := old[n-1]
	old[n-1] = nil
	node.index = -1
	*pq = old[0 : n-1]
	return node
}

func heuristic(p1, p2 Position) float64 {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func getNeighbors(pos Position, width, height int, obstacles map[Position]bool) []Position {
	neighbors := make([]Position, 0)
	directions := []Position{
		{X: 1, Y: 0},
		{X: -1, Y: 0},
		{X: 0, Y: 1},
		{X: 0, Y: -1},
		{X: 1, Y: 1},
		{X: -1, Y: 1},
		{X: 1, Y: -1},
		{X: -1, Y: -1},
	}

	for _, dir := range directions {
		newPos := Position{
			X: pos.X + dir.X,
			Y: pos.Y + dir.Y,
		}

		// Check bounds and obstacles
		if newPos.X >= 0 && newPos.X < float64(width) &&
			newPos.Y >= 0 && newPos.Y < float64(height) &&
			!obstacles[newPos] {
			neighbors = append(neighbors, newPos)
		}
	}

	return neighbors
}

func FindPath(start, goal Position, width, height int, obstacles map[Position]bool) []Position {
	openSet := &PriorityQueue{}
	heap.Init(openSet)

	startNode := &Node{
		Position: start,
		g:        0,
		h:        heuristic(start, goal),
		parent:   nil,
	}
	startNode.f = startNode.g + startNode.h

	heap.Push(openSet, startNode)
	closedSet := make(map[Position]bool)
	cameFrom := make(map[Position]*Node)

	for openSet.Len() > 0 {
		current := heap.Pop(openSet).(*Node)

		if current.Position == goal {
			path := make([]Position, 0)
			for node := current; node != nil; node = node.parent {
				path = append([]Position{node.Position}, path...)
			}
			return path
		}

		closedSet[current.Position] = true

		for _, neighbor := range getNeighbors(current.Position, width, height, obstacles) {
			if closedSet[neighbor] {
				continue
			}

			g := current.g + 1 // Cost to move to neighbor

			neighborNode := &Node{
				Position: neighbor,
				g:        g,
				h:        heuristic(neighbor, goal),
				parent:   current,
			}
			neighborNode.f = neighborNode.g + neighborNode.h

			heap.Push(openSet, neighborNode)
			cameFrom[neighbor] = current
		}
	}

	return nil // No path found
}
