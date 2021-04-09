package detrep

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (graph Graph) processResults(callback func(id string, pRank sdk.Uint, nRank sdk.Uint)) {
	graph.mergeNegatives()
	for key, node := range graph.Nodes {
		callback(key, node.PRank, node.NRank)
	}
}

func (graph Graph) mergeNegatives() {
	for key, node := range graph.NegNodes {
		// create the positive node if it doesn't exist
		if _, ok := graph.Nodes[node.ID]; ok == false {
			graph.Nodes[node.ID] = &Node{
				ID:       key,
				PRank:    sdk.ZeroUint(),
				degree:   sdk.ZeroUint(),
				nodeType: Positive,
			}
		}
		// write the neg rank value to the positive node
		graph.Nodes[node.ID].NRank = node.PRank
		// we don't want negative nodes around after they are merged
		delete(graph.Nodes, key)
	}
}
