// Package rep is an implementation of the Relevant Reputaiton protocol:
// a personalized pagerank algorithm that supports negative links
// personalized version offers sybil resistance
// can be used for voting, governance, ranking
// references:
// https://github.com/alixaxel/pagerank
// https://github.com/dcadenas/pagerank
// notes:
// Optimization - using int-indexed maps?
// TODO: edge case (only impacts display) - if a node has no inputs we should set its score to 0 to avoid
// a stale score if all of nodes inputs are cancelled out
// would need to keep track of node inputs...
package rep

import (
	"math"
	"strconv"
)

// MaxNegOffset defines the cutoff for when a node will have it's outging links counted
// if previously NegativeRank / PositiveRank > MaxNegOffset / (MaxNegOffset + 1) we will not consider
// any outgoing links
// otherwise, we counter the outgoing lings with one 'heavy' link proportional to the MaxNegOffset ratio
const MaxNegOffset = float64(10)

// NodeType is positive or negative
// each node in the graph can be represented by two nodes,
// a positive and a negative one
type NodeType int

// Positive nodes are consumers of positive links
// Negative nodes are consumers of neg links
const (
	Positive NodeType = iota
	Negative
)

// NodeInput is struct passing data to the graph
// TODO does it make sense to just use the Node type?
type NodeInput struct {
	ID    string
	PRank float64
	NRank float64
}

// Node is an internal node struct
type Node struct {
	id       string
	rank     float64 // pos page rank of the node
	rankNeg  float64 // only used when combining results
	degree   float64 // sum of all outgoing links
	nodeType NodeType
}

// Graph holds node and edge data.
type Graph struct {
	nodes       map[string]*Node
	negNodes    map[string]*Node
	edges       map[string](map[string]float64)
	params      RankParams
	negConsumer NodeInput
}

// RankParams is the pagerank parameters
// α is the probably the person will not teleport
// ε is the min global error between iterations
// personalization is the personalization vector (can be nil for non-personalized pr)
type RankParams struct {
	α, ε            float64
	personalization []string // array of ids
}

// NewGraph initializes and returns a new graph.
func NewGraph(α, ε, negConsumerRank float64) *Graph {
	return &Graph{
		nodes:    make(map[string]*Node),
		negNodes: make(map[string]*Node),
		edges:    make(map[string](map[string]float64)),
		params: RankParams{
			α:               α,
			ε:               ε,
			personalization: make([]string, 0),
		},
		negConsumer: NodeInput{ID: "negConsumer", PRank: negConsumerRank, NRank: 0},
	}
}

// NewNodeInput is ahelper method to create a node input struct
func NewNodeInput(id string, pRank float64, nRank float64) NodeInput {
	return NodeInput{ID: id, PRank: pRank, NRank: nRank}
}

// AddPersonalizationNode adds a node to the pagerank personlization vector
// these nodes will have high rank by default and all other rank will stem from them
// this makes non-personalaziation nodes sybil resistant (they cannot increase their own rank)
func (graph *Graph) AddPersonalizationNode(pNode NodeInput) {
	graph.params.personalization = append(graph.params.personalization, pNode.ID)
	// this to ensures source nodes exist
	graph.InitPosNode(pNode)
}

// Link creates a weighted edge between a source-target node pair.
// If the edge already exists, the weight is incremented.
func (graph *Graph) Link(source, target NodeInput, weight float64) {

	// if a node's neg/post rank ration is too high we don't process its links
	if source.PRank > 0 && source.NRank/source.PRank > MaxNegOffset/(MaxNegOffset+1) {
		return
	}

	sourceKey := getKey(source.ID, Positive)
	sourceNode := graph.initNode(sourceKey, source, Positive)

	// if weight is negative we use negative receiving node
	var nodeType NodeType
	if nodeType = Positive; weight < 0 {
		nodeType = Negative
	}
	targetKey := getKey(target.ID, nodeType)

	graph.initNode(targetKey, target, nodeType)

	sourceNode.degree += math.Abs(weight)

	if _, ok := graph.edges[sourceKey]; ok == false {
		graph.edges[sourceKey] = map[string]float64{}
	}
	graph.edges[sourceKey][targetKey] += math.Abs(weight)

	// note: use target.id here to make sure we reference the original id
	graph.cancelOpposites(*sourceNode, target.ID, nodeType)
}

// Finalize is the method that runs after all other inits and before pagerank
func (graph *Graph) Finalize() {
	graph.processNegatives()
}

// processNegatives creates an extra outgoing link from positive nodes
// if they have a negative counterpart
// this reduces the weight of the outgoing links from low-ranking nodes
func (graph *Graph) processNegatives() {
	negConsumerInput := graph.negConsumer

	for _, negNode := range graph.negNodes {
		// positive node doesn't exist
		if _, ok := graph.nodes[negNode.id]; ok == false {
			return
		}
		// node has no outgpoing links
		if graph.nodes[negNode.id].degree == 0 {
			return
		}

		posNode := graph.nodes[negNode.id]
		if posNode.rank == 0 {
			return
		}

		if negNode.rank >= posNode.rank {
			panic("negative ranking nodes should not have any degree") // this should never happen
		}

		negConsumer := graph.initNode(negConsumerInput.ID, negConsumerInput, Positive)

		var negMultiple float64

		// this first case should not happen because we ignore these links
		if negNode.rank/posNode.rank > MaxNegOffset/(MaxNegOffset+1) {
			// cap the degree multiple at MAX_NEG_OFFSET
			negMultiple = MaxNegOffset
		} else {
			negMultiple = 1/(1-negNode.rank/posNode.rank) - 1
		}

		// this is the weight we add to the outgoing node
		negWeight := negMultiple * graph.nodes[negNode.id].degree

		if _, ok := graph.edges[negNode.id]; ok == false {
			graph.edges[negNode.id] = map[string]float64{}
		}
		graph.edges[negNode.id][negConsumer.id] += negWeight
		graph.nodes[negNode.id].degree += negWeight
	}
}

// if there is both a positive and a negative link from A to B we cancel them out
func (graph *Graph) cancelOpposites(sourceNode Node, target string, nodeType NodeType) {
	key := getKey(target, nodeType)
	var oppositeKey string
	if oppositeKey = getKey(target, Positive); nodeType == Positive {
		oppositeKey = getKey(target, Negative)
	}

	if _, ok := graph.edges[sourceNode.id][oppositeKey]; ok == false {
		return
	}

	edge := graph.edges[sourceNode.id][key]
	opositeEdge := graph.edges[sourceNode.id][oppositeKey]

	switch {
	case opositeEdge > edge:
		graph.removeEdge(sourceNode.id, key)
		graph.edges[sourceNode.id][oppositeKey] -= edge
		// remove degree from both delete node and the adjustment
		sourceNode.degree -= 2 * edge

	case edge > opositeEdge:
		graph.removeEdge(sourceNode.id, oppositeKey)
		graph.edges[sourceNode.id][key] -= opositeEdge
		// remove degree from both delete node and the adjustment
		sourceNode.degree -= 2 * opositeEdge

	case edge == opositeEdge:
		graph.removeEdge(sourceNode.id, oppositeKey)
		graph.removeEdge(sourceNode.id, key)
		sourceNode.degree -= 2 * opositeEdge
	}
}

// InitPosNode is a helper method that initializes a positive node
func (graph *Graph) InitPosNode(inputNode NodeInput) *Node {
	return graph.initNode(inputNode.ID, inputNode, Positive)
}

// initNode initializes a node
func (graph *Graph) initNode(key string, inputNode NodeInput, nodeType NodeType) *Node {
	if _, ok := graph.nodes[key]; ok == false {
		graph.nodes[key] = &Node{
			id:       inputNode.ID, // id is independent of pos/neg keys
			degree:   0,
			nodeType: nodeType,
		}
		// store negative nodes so we can easily merge them later
		if nodeType == Negative {
			graph.negNodes[key] = graph.nodes[key]
		}
	}
	// update rank here in case we initilized with 0 early on
	var prevRank float64
	if prevRank = inputNode.PRank; nodeType == Negative {
		prevRank = inputNode.NRank
	}
	graph.nodes[key].rank = prevRank
	return graph.nodes[key]
}

// removeEdge removes edge from graph
func (graph *Graph) removeEdge(source string, target string) {
	delete(graph.edges[source], target)
	if len(graph.edges[source]) == 0 {
		delete(graph.edges, source)
	}
}

// getKey returns node id for positive nodes and <id>_neg for negative nodes
func getKey(key string, nodeType NodeType) string {
	if nodeType == Positive {
		return key
	}
	// TODO whats the best way to do this if we use int indexes? odd & even?
	return key + "_" + strconv.FormatInt(int64(nodeType), 10)
}
