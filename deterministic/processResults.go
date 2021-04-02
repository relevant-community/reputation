package detrep

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (graph Graph) processResults(callback func(id string, pRank sdk.Uint, nRank sdk.Uint)) {
	graph.mergeNegatives()
	for key, node := range graph.nodes {
		callback(key, node.PRank, node.NRank)
	}
}

func (graph Graph) mergeNegatives() {
	for key, node := range graph.negNodes {
		// create the positive node if it doesn't exist
		if _, ok := graph.nodes[node.ID]; ok == false {
			graph.nodes[node.ID] = &Node{
				ID:       key,
				PRank:    sdk.ZeroUint(),
				degree:   sdk.ZeroUint(),
				nodeType: Positive,
			}
		}
		// write the neg rank value to the positive node
		graph.nodes[node.ID].NRank = node.PRank
		// we don't want negative nodes around after they are merged
		delete(graph.nodes, key)
	}
}
