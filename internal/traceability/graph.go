package traceability

import (
	"fmt"
	"sort"
	"strings"
)

// NodeKind defines the canonical types of nodes in the traceability graph.
type NodeKind string

const (
	NodeRequirement NodeKind = "req"
	NodeDecision    NodeKind = "dec"
	NodeGoal        NodeKind = "goal"
	NodeCode        NodeKind = "code"
	NodeTest        NodeKind = "test"
	NodeDoc         NodeKind = "doc"
	NodeEvidence    NodeKind = "evidence"
	NodeExperiment  NodeKind = "experiment"
	NodeRejection   NodeKind = "rejection"
	NodeDebt        NodeKind = "debt"
)

func (k NodeKind) Valid() bool {
	switch k {
	case NodeRequirement, NodeDecision, NodeGoal, NodeCode, NodeTest,
		NodeDoc, NodeEvidence, NodeExperiment, NodeRejection, NodeDebt:
		return true
	}
	return false
}

// EdgeKind defines the typed relationship between graph nodes.
type EdgeKind string

const (
	EdgeSatisfies   EdgeKind = "satisfies"
	EdgeDerivesFrom EdgeKind = "derives_from"
	EdgeImplements  EdgeKind = "implements"
	EdgeVerifies    EdgeKind = "verifies"
	EdgeDocuments   EdgeKind = "documents"
	EdgeEvidencedBy EdgeKind = "evidenced_by"
	EdgeRejects     EdgeKind = "rejects"
	EdgeIncursDebt  EdgeKind = "incurs_debt"
)

func (k EdgeKind) Valid() bool {
	switch k {
	case EdgeSatisfies, EdgeDerivesFrom, EdgeImplements, EdgeVerifies,
		EdgeDocuments, EdgeEvidencedBy, EdgeRejects, EdgeIncursDebt:
		return true
	}
	return false
}

// Node represents an element in the traceability graph.
type Node struct {
	ID       string            `json:"id"`
	Kind     NodeKind          `json:"kind"`
	Title    string            `json:"title"`
	Ref      string            `json:"ref,omitempty"` // file path, goal ID, or contract ID
	Metadata map[string]string `json:"metadata,omitempty"`
}

// TracePath captures the bidirectional traversal results for a given node reference.
type TracePath struct {
	Root        Node   `json:"root"`
	Upstream    []Node `json:"upstream"`
	Downstream  []Node `json:"downstream"`
	LinkedEdges []Edge `json:"linked_edges"`
}

// Graph models the typed traceability network (requirements ↔ decisions ↔ code ↔ tests ↔ docs ↔ evidence).
type Graph struct {
	Nodes map[string]Node `json:"nodes"`
	Edges []Edge          `json:"edges"`
}

// NewGraph initializes an empty traceability graph.
func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[string]Node),
		Edges: make([]Edge, 0),
	}
}

// AddNode inserts a node into the graph after validation.
func (g *Graph) AddNode(n Node) error {
	if strings.TrimSpace(n.ID) == "" {
		return fmt.Errorf("node requires a non-empty ID")
	}
	if !n.Kind.Valid() {
		return fmt.Errorf("invalid node kind %q", n.Kind)
	}
	if g.Nodes == nil {
		g.Nodes = make(map[string]Node)
	}
	g.Nodes[n.ID] = n
	return nil
}

// AddEdge inserts a directed typed edge into the graph.
func (g *Graph) AddEdge(e Edge) error {
	if strings.TrimSpace(e.ID) == "" {
		e.ID = fmt.Sprintf("%s-%s-%s", e.From, e.Kind, e.To)
	}
	if _, ok := g.Nodes[e.From]; !ok {
		return fmt.Errorf("source node %q does not exist in graph", e.From)
	}
	if _, ok := g.Nodes[e.To]; !ok {
		return fmt.Errorf("target node %q does not exist in graph", e.To)
	}
	edgeKind := EdgeKind(e.Kind)
	if !edgeKind.Valid() {
		return fmt.Errorf("invalid edge kind %q", e.Kind)
	}
	g.Edges = append(g.Edges, e)
	return nil
}

// FindNode finds a node by exact ID or matching Ref.
func (g *Graph) FindNode(query string) (Node, bool) {
	if n, ok := g.Nodes[query]; ok {
		return n, true
	}
	for _, n := range g.Nodes {
		if n.Ref == query || strings.EqualFold(n.Title, query) {
			return n, true
		}
	}
	return Node{}, false
}

// Trace computes the upstream and downstream lineages for a given target reference.
func (g *Graph) Trace(ref string) (TracePath, error) {
	root, ok := g.FindNode(ref)
	if !ok {
		return TracePath{}, fmt.Errorf("trace node %q not found in graph", ref)
	}

	result := TracePath{
		Root:        root,
		Upstream:    make([]Node, 0),
		Downstream:  make([]Node, 0),
		LinkedEdges: make([]Edge, 0),
	}

	visitedUp := make(map[string]bool)
	visitedDown := make(map[string]bool)

	// Traverse upstream (towards requirements/decisions: following From == curr -> To)
	var qUp []string
	qUp = append(qUp, root.ID)
	for len(qUp) > 0 {
		curr := qUp[0]
		qUp = qUp[1:]

		for _, e := range g.Edges {
			if e.From == curr && !visitedUp[e.To] {
				visitedUp[e.To] = true
				if n, exists := g.Nodes[e.To]; exists {
					result.Upstream = append(result.Upstream, n)
					result.LinkedEdges = append(result.LinkedEdges, e)
					qUp = append(qUp, e.To)
				}
			}
		}
	}

	// Traverse downstream (towards verification/evidence: following To == curr -> From)
	var qDown []string
	qDown = append(qDown, root.ID)
	for len(qDown) > 0 {
		curr := qDown[0]
		qDown = qDown[1:]

		for _, e := range g.Edges {
			if e.To == curr && !visitedDown[e.From] {
				visitedDown[e.From] = true
				if n, exists := g.Nodes[e.From]; exists {
					result.Downstream = append(result.Downstream, n)
					result.LinkedEdges = append(result.LinkedEdges, e)
					qDown = append(qDown, e.From)
				}
			}
		}
	}

	// Stable sort for deterministic presentation
	sort.Slice(result.Upstream, func(i, j int) bool { return result.Upstream[i].ID < result.Upstream[j].ID })
	sort.Slice(result.Downstream, func(i, j int) bool { return result.Downstream[i].ID < result.Downstream[j].ID })

	return result, nil
}
