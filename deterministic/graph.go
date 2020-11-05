// Package reputation is an implementation of the Relevant Reputaiton protocol:
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
package reputation_det

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

const Decimals = 18

// this is defines the cutoff for when a node will have it's outging links counted
// if previously NegativeRank / PositiveRank > MaxNegOffset / (MaxNegOffset + 1) we will not consider
// any outgoing links
// otherwise, we counter the outgoing lings with one 'heavy' link proportional to the MaxNegOffset ratio
const MaxNegOffset = 10

type NodeInput struct {
	Id    string
	PRank sdk.Uint
	NRank sdk.Uint
}

type Node struct {
	id       string
	rank     sdk.Uint // pos page rank of the node
	rankNeg  sdk.Uint // only used when combining results
	degree   sdk.Uint // sum of all outgoing links
	nodeType NodeType
}

// Graph holds node and edge data.
type Graph struct {
	nodes        map[string]*Node
	negNodes     map[string]*Node
	edges        map[string](map[string]sdk.Uint)
	params       RankParams
	negConsumer  NodeInput
	Precision    sdk.Uint
	MaxNegOffset sdk.Uint
}

type RankParams struct {
	α, ε            sdk.Uint
	personalization []string
}

// NewGraph initializes and returns a new graph.
func NewGraph(α sdk.Uint, ε sdk.Uint, negConsumerRank sdk.Uint) *Graph {
	return &Graph{
		nodes:    make(map[string]*Node),
		negNodes: make(map[string]*Node),
		edges:    make(map[string](map[string]sdk.Uint)),
		params: RankParams{
			α:               α,
			ε:               ε,
			personalization: make([]string, 0),
		},
		negConsumer:  NodeInput{Id: "negConsumer", PRank: negConsumerRank, NRank: sdk.ZeroUint()},
		Precision:    sdk.NewUintFromBigInt(sdk.NewIntWithDecimal(1, Decimals).BigInt()),
		MaxNegOffset: sdk.NewUintFromBigInt(sdk.NewIntWithDecimal(MaxNegOffset, Decimals).BigInt()),
	}
}

// helper method to create a node input struct
func NewNodeInput(id string, pRank sdk.Uint, nRank sdk.Uint) NodeInput {
	return NodeInput{Id: id, PRank: pRank, NRank: nRank}
}

// AddPersonalizationNode adds a node to the pagerank personlization vector
// these nodes will have high rank by default and all other rank will stem from them
// this makes non-personalaziation nodes sybil resistant (they cannot increase their own rank)
func (graph *Graph) AddPersonalizationNode(pNode NodeInput) {
	graph.params.personalization = append(graph.params.personalization, pNode.Id)
	// this to ensures source nodes exist
	graph.InitPosNode(pNode)
}

// Link creates a weighted edge between a source-target node pair.
// If the edge already exists, the weight is incremented.
func (graph *Graph) Link(source, target NodeInput, weight sdk.Int) {

	// if a node's neg/pos rank is > MaxNegOffset / (MaxNegOffset + 1) we don't process it
	if source.PRank.GT(sdk.ZeroUint()) {
		negPosRatio := source.NRank.Mul(graph.Precision).Quo(source.PRank)
		one := graph.Precision
		if negPosRatio.GT(graph.MaxNegOffset.Mul(graph.Precision).Quo(graph.MaxNegOffset.Add(one))) {
			return
		}
	}

	sourceKey := getKey(source.Id, Positive)
	sourceNode := graph.initNode(sourceKey, source, Positive)

	// if weight is negative we use negative receiving node
	var nodeType NodeType
	var weightUint sdk.Uint
	if weight.LT(sdk.ZeroInt()) {
		nodeType = Negative
		weightUint = sdk.NewUintFromBigInt(weight.Neg().BigInt())
	} else {
		nodeType = Positive
		weightUint = sdk.NewUintFromBigInt(weight.BigInt())
	}
	targetKey := getKey(target.Id, nodeType)

	graph.initNode(targetKey, target, nodeType)

	sourceNode.degree = sourceNode.degree.Add(weightUint)

	if _, ok := graph.edges[sourceKey]; ok == false {
		graph.edges[sourceKey] = map[string]sdk.Uint{}
	}

	if _, ok := graph.edges[sourceKey][targetKey]; ok == false {
		graph.edges[sourceKey][targetKey] = sdk.ZeroUint()
	}

	graph.edges[sourceKey][targetKey] = graph.edges[sourceKey][targetKey].Add(weightUint)

	// note: use target.id here to make sure we reference the original id
	graph.cancelOpposites(*sourceNode, target.Id, nodeType)
}

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
		if graph.nodes[negNode.id].degree.IsZero() {
			return
		}

		posNode := graph.nodes[negNode.id]
		if posNode.rank.IsZero() || negNode.rank.IsZero() {
			return
		}
		negConsumer := graph.initNode(negConsumerInput.Id, negConsumerInput, Positive)

		one := graph.Precision

		// posNode.rank is not 0 check above
		negPosRatio := negNode.rank.Mul(graph.Precision).Quo(posNode.rank)

		var negMultiple sdk.Uint

		// if negPosRatio > MaxNegOffset / (MaxNegOffset + 1) we use the MaxNegOffset
		// this first case should not happen because we ignore these links
		if negPosRatio.GT(graph.MaxNegOffset.Mul(graph.Precision).Quo(graph.MaxNegOffset.Add(one))) {
			negMultiple = graph.MaxNegOffset
		} else {
			denom := one.Sub(negPosRatio)
			negMultiple = one.Mul(graph.Precision).Quo(denom).Sub(one)
		}
		// cap the vote decrease at 10x
		negWeight := negMultiple.Mul(graph.nodes[negNode.id].degree).Quo(graph.Precision)

		// this should actually never happen if degree is > 0
		if _, ok := graph.edges[negNode.id]; ok == false {
			graph.edges[negNode.id] = map[string]sdk.Uint{}
		}

		if _, ok := graph.edges[negNode.id][negConsumer.id]; ok == false {
			graph.edges[negNode.id][negConsumer.id] = sdk.ZeroUint()
		}

		graph.edges[negNode.id][negConsumer.id] = graph.edges[negNode.id][negConsumer.id].Add(negWeight)
		graph.nodes[negNode.id].degree = graph.nodes[negNode.id].degree.Add(negWeight)
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
	case opositeEdge.GT(edge):
		graph.removeEdge(sourceNode.id, key)
		graph.edges[sourceNode.id][oppositeKey] = opositeEdge.Sub(edge)
		// remove degree from both delete node and the adjustment
		sourceNode.degree = sourceNode.degree.Sub(edge.Mul(sdk.NewUint(2)))

	case edge.GT(opositeEdge):
		graph.removeEdge(sourceNode.id, oppositeKey)
		graph.edges[sourceNode.id][key] = edge.Sub(opositeEdge)
		// remove degree from both delete node and the adjustment
		sourceNode.degree = sourceNode.degree.Sub(opositeEdge.Mul(sdk.NewUint(2)))

	case edge.Equal(opositeEdge):
		graph.removeEdge(sourceNode.id, oppositeKey)
		graph.removeEdge(sourceNode.id, key)

		sourceNode.degree = sourceNode.degree.Sub(opositeEdge.Mul(sdk.NewUint(2)))
	}
}

func (graph *Graph) InitPosNode(inputNode NodeInput) *Node {
	return graph.initNode(inputNode.Id, inputNode, Positive)
}

func (graph *Graph) initNode(key string, inputNode NodeInput, nodeType NodeType) *Node {
	if _, ok := graph.nodes[key]; ok == false {
		graph.nodes[key] = &Node{
			id:       inputNode.Id, // id is independent of pos/neg keys
			degree:   sdk.ZeroUint(),
			rank:     sdk.ZeroUint(),
			rankNeg:  sdk.ZeroUint(),
			nodeType: nodeType,
		}
		// store negative nodes so we can easily merge them later
		if nodeType == Negative {
			graph.negNodes[key] = graph.nodes[key]
		}
	}
	// update rank here in case we initilized with 0 early on
	var prevRank sdk.Uint
	if prevRank = inputNode.PRank; nodeType == Negative {
		prevRank = inputNode.NRank
	}
	graph.nodes[key].rank = prevRank
	return graph.nodes[key]
}

func (graph *Graph) removeEdge(source string, target string) {
	delete(graph.edges[source], target)
	if len(graph.edges[source]) == 0 {
		delete(graph.edges, source)
	}
}

func getKey(key string, nodeType NodeType) string {
	if nodeType == Positive {
		return key
	}
	// TODO whats the best way to do this if we use int indexes? odd & even?
	return key + "_" + strconv.FormatInt(int64(nodeType), 10)
}
