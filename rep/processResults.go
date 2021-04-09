package rep

func (graph Graph) processResults(callback func(id string, pRank float64, nRank float64)) {
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
				PRank:    0,
				degree:   0,
				nodeType: Positive,
			}
		}
		// write the neg rank value to the positive node
		graph.Nodes[node.ID].NRank = node.PRank
		// we don't want negative nodes around after they are merged
		delete(graph.Nodes, key)
	}
}
