package models

import (
	"container/heap"
	"fmt"
	"math"
	"math/rand"
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
			newPos.Y >= 0 && newPos.Y < float64(height) {
			// Convert to integer position for obstacle checking
			intPos := Position{
				X: math.Floor(newPos.X),
				Y: math.Floor(newPos.Y),
			}
			if !obstacles[intPos] {
				neighbors = append(neighbors, newPos)
				//fmt.Printf("Valid neighbor found: {%.2f, %.2f}\n", newPos.X, newPos.Y)
			}
		}
	}

	return neighbors
}

func ConvertPathToFloat(intPath []Position) []Position {
	floatPath := make([]Position, len(intPath))
	for i, pos := range intPath {
		// Add small random offset to avoid collisions
		// but stay within safe bounds of the cell
		offsetX := 0.2 + rand.Float64()*0.6
		offsetY := 0.2 + rand.Float64()*0.6

		floatPath[i] = Position{
			X: pos.X + offsetX,
			Y: pos.Y + offsetY,
		}

		floatPath[i] = floatPath[i].Round()
	}
	return floatPath
}

func FindPath(start, goal Position, width, height int, obstacles map[Position]bool) []Position {
	// Convert to integer coordinates for pathfinding
	startInt := Position{
		X: math.Floor(start.X),
		Y: math.Floor(start.Y),
	}
	goalInt := Position{
		X: math.Floor(goal.X),
		Y: math.Floor(goal.Y),
	}

	fmt.Printf("\nFindPath starting: from {%.2f, %.2f} to {%.2f, %.2f}\n",
		startInt.X, startInt.Y, goalInt.X, goalInt.Y)

	openSet := &PriorityQueue{}
	heap.Init(openSet)

	startNode := &Node{
		Position: startInt,
		g:        0,
		h:        heuristic(startInt, goalInt),
		parent:   nil,
	}
	startNode.f = startNode.g + startNode.h

	openNodes := make(map[Position]*Node)
	openNodes[startInt] = startNode

	heap.Push(openSet, startNode)
	closedSet := make(map[Position]bool)

	maxIterations := width * height * 2
	iterCount := 0

	for openSet.Len() > 0 {
		iterCount++
		if iterCount > maxIterations {
			fmt.Printf("Path search exceeded max iterations (%d), aborting\n", maxIterations)
			return nil
		}

		current := heap.Pop(openSet).(*Node)
		delete(openNodes, current.Position)

		if current.Position == goalInt {
			// Found the path - convert to float coordinates
			intPath := make([]Position, 0)
			for node := current; node != nil; node = node.parent {
				intPath = append([]Position{node.Position}, intPath...)
			}
			floatPath := ConvertPathToFloat(intPath)
			fmt.Printf("Path found! Length: %d\n", len(floatPath))
			return floatPath
		}

		closedSet[current.Position] = true
		neighbors := getNeighbors(current.Position, width, height, obstacles)

		for _, neighbor := range neighbors {
			// Convert neighbor to integer position for comparisons
			neighborInt := Position{
				X: math.Floor(neighbor.X),
				Y: math.Floor(neighbor.Y),
			}

			if closedSet[neighborInt] {
				continue
			}

			g := current.g + 1

			if existingNode, exists := openNodes[neighborInt]; exists {
				if g < existingNode.g {
					existingNode.g = g
					existingNode.f = g + existingNode.h
					existingNode.parent = current
				}
				continue
			}

			neighborNode := &Node{
				Position: neighborInt,
				g:        g,
				h:        heuristic(neighborInt, goalInt),
				parent:   current,
			}
			neighborNode.f = neighborNode.g + neighborNode.h

			heap.Push(openSet, neighborNode)
			openNodes[neighborInt] = neighborNode
		}
	}

	fmt.Printf("No path found after %d iterations\n", iterCount)
	return nil
}
