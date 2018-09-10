package main

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/twmb/algoimpl/go/graph"
)

func makeGraph() (g *graph.Graph, node1, node2, node3, node4, node5 *graph.Node) {
	rand.Seed(math.MaxInt64)

	A := EximchainNode{"a"}
	B := EximchainNode{"b"}
	C := EximchainNode{"c"}
	D := EximchainNode{"d"}
	E := EximchainNode{"e"}
	g = graph.New(graph.Undirected)

	na := g.MakeNode()
	na.Value = A.NodeToIntfAddr()

	nb := g.MakeNode()
	nb.Value = B.NodeToIntfAddr()

	nc := g.MakeNode()
	nc.Value = C.NodeToIntfAddr()

	nd := g.MakeNode()
	nd.Value = D.NodeToIntfAddr()

	ne := g.MakeNode()
	ne.Value = E.NodeToIntfAddr()

	node1 = na
	node2 = nb
	node3 = nc
	node4 = nd
	node5 = ne

	return
}

func ngMakeCut1() (result int) {
	g, na, nb, nc, nd, ne := makeGraph()
	g.MakeEdge(na, nb)
	g.MakeEdge(nb, nc)
	g.MakeEdge(nc, na)
	g.MakeEdge(nd, ne)

	g.MakeEdge(nb, nd)

	cut1 := g.RandMinimumCut(100, 1)
	result = len(cut1)
	return
}

func ngMakeCut2() (result int) {
	g, na, nb, nc, nd, ne := makeGraph()
	g.MakeEdge(na, nb)
	g.MakeEdge(nb, nc)
	g.MakeEdge(nc, na)
	g.MakeEdge(nd, ne)

	g.MakeEdge(nb, nd)
	g.MakeEdge(nc, ne)

	cut1 := g.RandMinimumCut(100, 1)
	result = len(cut1)
	return
}

func makeCut1() (result int) {
	rand.Seed(math.MaxInt64)

	A := EximchainNode{"a"}
	B := EximchainNode{"b"}
	C := EximchainNode{"c"}
	D := EximchainNode{"d"}
	E := EximchainNode{"e"}
	g := graph.New(graph.Undirected)

	na := g.MakeNode()
	na.Value = A.NodeToIntfAddr()

	nb := g.MakeNode()
	nb.Value = B.NodeToIntfAddr()

	nc := g.MakeNode()
	nc.Value = C.NodeToIntfAddr()

	nd := g.MakeNode()
	nd.Value = D.NodeToIntfAddr()

	ne := g.MakeNode()
	ne.Value = E.NodeToIntfAddr()

	g.MakeEdge(na, nb)
	g.MakeEdge(nb, nc)
	g.MakeEdge(nc, na)
	g.MakeEdge(nd, ne)

	g.MakeEdge(nb, nd)

	cut1 := g.RandMinimumCut(100, 1)
	result = len(cut1)

	return
}

func makeCut2() (result int) {
	rand.Seed(math.MaxInt64)

	A := EximchainNode{"a"}
	B := EximchainNode{"b"}
	C := EximchainNode{"c"}
	D := EximchainNode{"d"}
	E := EximchainNode{"e"}
	g := graph.New(graph.Undirected)

	na := g.MakeNode()
	na.Value = A.NodeToIntfAddr()

	nb := g.MakeNode()
	nb.Value = B.NodeToIntfAddr()

	nc := g.MakeNode()
	nc.Value = C.NodeToIntfAddr()

	nd := g.MakeNode()
	nd.Value = D.NodeToIntfAddr()

	ne := g.MakeNode()
	ne.Value = E.NodeToIntfAddr()

	g.MakeEdge(na, nb)
	g.MakeEdge(nb, nc)
	g.MakeEdge(nc, na)
	g.MakeEdge(nd, ne)

	g.MakeEdge(nb, nd)
	g.MakeEdge(nc, ne)

	subgraphs := g.StronglyConnectedComponents()
	for _, subgraph := range subgraphs {
		for _, node := range subgraph {
			eximchainNode := IntfToEximchainNode(node.Value)
			if eximchainNode == nil || eximchainNode.IP == "" {
				panic("Eximchain Node has no IP!")
			}
		}

	}

	cut2 := g.RandMinimumCut(100, 1)
	result = len(cut2)

	return
}

func TestMakeCut1(t *testing.T) {
	if res := makeCut1(); res != 1 {
		t.Fatalf("Expected MakeCut1 to return 1, but got %d", res)
	}
}

func TestMakeCut2(t *testing.T) {
	if res := makeCut2(); res != 2 {
		t.Fatalf("Expected MakeCut1 to return 2, but got %d", res)
	}
}

// Based on https://math.stackexchange.com/questions/240556/radius-diameter-and-center-of-graph
func TestDiameter1(t *testing.T) {
	g := NewGraphContainer(graph.Undirected)
	a1 := g.MakeNode()
	b1 := g.MakeNode()
	c1 := g.MakeNode()
	d1 := g.MakeNode()
	e1 := g.MakeNode()
	f1 := g.MakeNode()
	g1 := g.MakeNode()
	h1 := g.MakeNode()
	i1 := g.MakeNode()
	j1 := g.MakeNode()
	k1 := g.MakeNode()
	l1 := g.MakeNode()
	m1 := g.MakeNode()

	g.MakeEdge(a1, b1) // 1
	g.MakeEdge(b1, c1) // 2
	g.MakeEdge(b1, d1) // 3
	g.MakeEdge(d1, e1) // 4
	g.MakeEdge(d1, c1) // 5
	g.MakeEdge(d1, f1) // 6
	g.MakeEdge(e1, f1) // 7
	g.MakeEdge(f1, g1) // 8
	g.MakeEdge(f1, h1) // 9
	g.MakeEdge(f1, i1) // 10
	g.MakeEdge(i1, j1) // 11
	g.MakeEdge(i1, h1) // 12
	g.MakeEdge(i1, k1) // 13
	g.MakeEdge(i1, l1) // 14
	g.MakeEdge(k1, l1) // 15
	g.MakeEdge(l1, m1) // 16

	expectedDiameter := 6
	if diameter := g.Diameter(); diameter != expectedDiameter {
		t.Fatalf("Expected diameter is %d, but got %d!", expectedDiameter, diameter)
	}

}

func Graph1() (result *GraphContainer) {
	ag := NewGraphContainer(graph.Undirected)
	a := ag.MakeNode()
	b := ag.MakeNode()
	c := ag.MakeNode()
	d := ag.MakeNode()
	e := ag.MakeNode()
	f := ag.MakeNode()
	g := ag.MakeNode()

	ag.MakeEdge(a, b)
	ag.MakeEdge(b, c)
	ag.MakeEdge(b, d)
	ag.MakeEdge(b, e)
	ag.MakeEdge(d, f)
	ag.MakeEdge(e, g)

	result = ag
	return
}

func Graph2() (result *GraphContainer) {
	ag := NewGraphContainer(graph.Undirected)
	a := ag.MakeNode()
	b := ag.MakeNode()
	c := ag.MakeNode()
	d := ag.MakeNode()
	e := ag.MakeNode()
	f := ag.MakeNode()
	g := ag.MakeNode()

	ag.MakeEdge(a, b)
	ag.MakeEdge(b, c)
	ag.MakeEdge(b, d)
	ag.MakeEdge(b, e)
	ag.MakeEdge(b, f)
	ag.MakeEdge(f, g)

	result = ag
	return
}

// Based on https://www.researchgate.net/figure/Example-of-Graph-diameter-and-radius_fig1_280012327
func TestDiameter2(t *testing.T) {
	ag := Graph1()
	assertEqual(t, ag.Diameter(), 4, "")
}

// Based on https://www.researchgate.net/figure/Example-of-Graph-diameter-and-radius_fig1_280012327
func TestDiameter3(t *testing.T) {
	ag := Graph2()
	assertEqual(t, ag.Diameter(), 3, "")
}

// Based on https://www.researchgate.net/figure/Example-of-Graph-diameter-and-radius_fig1_280012327
func TestRadius1(t *testing.T) {
	g := Graph1()
	assertEqual(t, g.Radius(), 2, "")
}

// Based on https://www.researchgate.net/figure/Example-of-Graph-diameter-and-radius_fig1_280012327
func TestRadius2(t *testing.T) {
	g := Graph2()
	assertEqual(t, g.Radius(), 2, "")
}

func assertEqual(t *testing.T, actual interface{}, expected interface{}, message string) {
	if expected == actual {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("expected %v but got %v", expected, actual)
	}
	t.Fatal(message)
}

// Based on https://www.tutorialspoint.com/graph_theory/graph_theory_basic_properties.htm
func TestDistance(t *testing.T) {
	ag := NewGraphContainer(graph.Undirected)
	a := ag.MakeNamedNode("a")
	b := ag.MakeNamedNode("b")
	c := ag.MakeNamedNode("c")
	d := ag.MakeNamedNode("d")
	e := ag.MakeNamedNode("e")
	f := ag.MakeNamedNode("f")
	g := ag.MakeNamedNode("g")

	ag.MakeEdge(a, b)
	ag.MakeEdge(a, c)
	ag.MakeEdge(a, d)
	ag.MakeEdge(b, e)
	ag.MakeEdge(c, f)
	ag.MakeEdge(d, e)
	ag.MakeEdge(d, f)
	ag.MakeEdge(e, g)
	ag.MakeEdge(f, g)

	assertEqual(t, ag.Distance(a, b), 1, "")
	assertEqual(t, ag.Distance(a, c), 1, "")
	assertEqual(t, ag.Distance(a, d), 1, "")
	assertEqual(t, ag.Distance(a, e), 2, "")
	assertEqual(t, ag.Distance(a, f), 2, "")
	assertEqual(t, ag.Distance(a, g), 3, "")
}

// Based on https://www.tutorialspoint.com/graph_theory/graph_theory_basic_properties.htm
func TestEccentricity(t *testing.T) {
	ag := NewGraphContainer(graph.Undirected)
	a := ag.MakeNode()
	b := ag.MakeNode()
	c := ag.MakeNode()
	d := ag.MakeNode()
	e := ag.MakeNode()
	f := ag.MakeNode()
	g := ag.MakeNode()

	ag.MakeEdge(a, b)
	ag.MakeEdge(a, c)
	ag.MakeEdge(a, d)
	ag.MakeEdge(b, e)
	ag.MakeEdge(c, f)
	ag.MakeEdge(d, e)
	ag.MakeEdge(d, f)
	ag.MakeEdge(e, g)
	ag.MakeEdge(f, g)

	assertEqual(t, ag.Eccentricity(a), 3, "")
	assertEqual(t, ag.Eccentricity(b), 3, "")
	assertEqual(t, ag.Eccentricity(c), 3, "")
	assertEqual(t, ag.Eccentricity(d), 2, "")
	assertEqual(t, ag.Eccentricity(e), 3, "")
	assertEqual(t, ag.Eccentricity(f), 3, "")
	assertEqual(t, ag.Eccentricity(g), 3, "")

}

func TestRadius3(t *testing.T) {
	ag := NewGraphContainer(graph.Undirected)
	a := ag.MakeNode()
	b := ag.MakeNode()
	c := ag.MakeNode()
	d := ag.MakeNode()
	e := ag.MakeNode()
	f := ag.MakeNode()
	g := ag.MakeNode()

	ag.MakeEdge(a, b)
	ag.MakeEdge(a, c)
	ag.MakeEdge(a, d)
	ag.MakeEdge(b, e)
	ag.MakeEdge(c, f)
	ag.MakeEdge(d, e)
	ag.MakeEdge(d, f)
	ag.MakeEdge(e, g)
	ag.MakeEdge(f, g)

	assertEqual(t, ag.Radius(), 2, "")
}
