package reputation_det

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Results struct {
	rankPos float64
	rankNeg float64
	id      string
}

func (graph Graph) processResults(callback func(id string, pRank sdk.Uint, nRank sdk.Uint)) {
	graph.mergeNegatives()
	for key, node := range graph.nodes {
		callback(key, node.rank, node.rankNeg)
	}
}

func (graph Graph) mergeNegatives() {
	for key, node := range graph.negNodes {
		// create the positive node if it doesn't exist
		if _, ok := graph.nodes[node.id]; ok == false {
			graph.nodes[node.id] = &Node{
				id:       key,
				rank:     sdk.ZeroUint(),
				degree:   sdk.ZeroUint(),
				nodeType: Positive,
			}
		}
		// write the neg rank value to the positive node
		graph.nodes[node.id].rankNeg = node.rank
		// we don't want negative nodes around after they are merged
		delete(graph.nodes, key)
	}
}
