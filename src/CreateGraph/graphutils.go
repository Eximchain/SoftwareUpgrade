package main

import (
	"math"

	"github.com/twmb/algoimpl/go/graph"
)

type (
	GraphContainer struct {
		g     *graph.Graph
		nodes []*graph.Node
	}
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// NewGraphContainer creates a container for the graph as well as the graph nodes
func NewGraphContainer(kind graph.GraphType) (result *GraphContainer) {
	result = &GraphContainer{
		graph.New(kind),
		make([]*graph.Node, 0)}
	return result
}

// AllPathSearch calls the graph's AllPathSearch method
func (container *GraphContainer) AllPathSearch(a, b *graph.Node) []graph.Path {
	return container.g.AllPathSearch(a, b)
}

// Diameter calculates the diameter of te graph by taking the maximum length of the
// Dijkstra path of all the nodes. If nodes are removed, this app will need to keep track of it too, otherwise, the results
// will be unexpected. See container.nodes
func (container *GraphContainer) Diameter() (result int) {
	paths := make([][]graph.Path, len(container.nodes))
	result = 0
	for i := 0; i < len(paths); i++ {
		paths[i] = container.g.DijkstraSearch(*container.nodes[i])
		for j := 0; j < len(paths[i]); j++ {
			result = max(result, len(paths[i][j].Path))
		}
	}
	return
}

// Distance calculates all the paths between node1 and node2 and returns
// the minimum of all the paths.
func (container *GraphContainer) Distance(node1, node2 *graph.Node) (result int) {
	result = math.MaxInt16
	paths := container.g.AllPathSearch(node1, node2)
	for _, path := range paths {
		result = min(result, len(path.Path))
	}
	return
}

// Eccentricity calculates all the distances between other nodes and the given node
// It then takes the maximum of all the distances and returns it.
func (container *GraphContainer) Eccentricity(node *graph.Node) (result int) {
	distances := make([]int, len(container.nodes)-1)
	j := 0
	for i := 0; i < len(container.nodes); i++ {
		if container.nodes[i] == node {
			continue
		}
		distances[j] = container.Distance(node, container.nodes[i])
		result = max(result, distances[j])
		j++
	}
	return
}

// Radius calculates the radius of the graph
func (container *GraphContainer) Radius() (result int) {
	result = math.MaxInt16
	eccentricities := make([]int, len(container.nodes))
	for i := 0; i < len(container.nodes); i++ {
		eccentricities[i] = container.Eccentricity(container.nodes[i])
		result = min(result, eccentricities[i])
	}
	return
}

// MakeEdge is a wrapper to the graph's MakeEdge function
func (container *GraphContainer) MakeEdge(from, to *graph.Node) (result error) {
	return container.g.MakeEdge(from, to)
}

// MakeNode is a wrapper to the graph's MakeNode function
func (container *GraphContainer) MakeNode() (result *graph.Node) {
	result = container.g.MakeNode()
	container.nodes = append(container.nodes, result)
	return
}

// MakeNamedNode calls the graph's MakeNamedNode function and keeps a copy of it
func (container *GraphContainer) MakeNamedNode(name string) (result *graph.Node) {
	result = container.g.MakeNamedNode(name)
	container.nodes = append(container.nodes, result)
	return
}

// RandMinimumCut calls the graph's RandMinimumCut and returns the result
func (container *GraphContainer) RandMinimumCut(iterations, concurrent int) (result []graph.Edge) {
	result = container.g.RandMinimumCut(iterations, concurrent)
	return
}

// StronglyConnectedComponents calls the graph's StronglyConnectedComponents and returns the result
func (container *GraphContainer) StronglyConnectedComponents() (result [][]graph.Node) {
	result = container.g.StronglyConnectedComponents()
	return
}
