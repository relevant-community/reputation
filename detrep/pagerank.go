package detrep

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Rank computes the PageRank of every node in the directed graph.
// α (alpha) is the damping factor, usually set to 0.85.
// ε (epsilon) is the convergence criteria, usually set to a tiny value.
//
// This method will run as many iterations as needed, until the graph converges.
func (graph Graph) Rank(callback func(key string, pRank sdk.Uint, nRank sdk.Uint)) {
	graph.Finalize()

	one := graph.Precision

	Δ := one
	N := sdk.NewUint(uint64(len(graph.Nodes)))
	pVector := graph.Params.Personalization
	ε := graph.Params.ε
	α := graph.Params.α

	personalized := len(pVector) > 0

	// these are personlaization node weights
	// we adjust them so that all p nodes have the same outgoing link weight
	pWeights := graph.initPersonalizationNodes()

	// Normalize all the edge weights so that their sum amounts to 1.
	for source := range graph.Nodes {
		if graph.Nodes[source].degree.GT(sdk.ZeroUint()) {
			for target := range graph.Edges[source] {
				graph.Edges[source][target] = graph.Edges[source][target].Mul(graph.Precision).Quo(graph.Nodes[source].degree)
			}
		}
	}

	graph.initScores(N, pWeights)

	iter := 0
	for Δ.GT(ε) {
		danglingWeight := sdk.ZeroUint()
		nodes := map[string]sdk.Uint{}

		for key, value := range graph.Nodes {
			nodes[key] = value.PRank

			if value.degree.IsZero() {
				danglingWeight = danglingWeight.Add(value.PRank)
			}

			graph.Nodes[key].PRank = sdk.ZeroUint()
		}

		danglingWeight = danglingWeight.Mul(α).Quo(graph.Precision)

		for source := range graph.Nodes {
			for target, weight := range graph.Edges[source] {
				addWeight := α.Mul(nodes[source]).Quo(graph.Precision).Mul(weight).Quo(graph.Precision)
				graph.Nodes[target].PRank = graph.Nodes[target].PRank.Add(addWeight)
			}

			if !personalized {
				graph.Nodes[source].PRank = graph.Nodes[source].PRank.Add(one.Sub(α).Quo(N).Add(danglingWeight.Quo(N)))
			}
		}

		// random jump + dangling weights are transferred to admins
		// this makes pagerank sybil resistant
		if personalized {
			for i, root := range pVector {
				graph.Nodes[root].PRank = graph.Nodes[root].PRank.Add((one.Sub(α).Add(danglingWeight)).Mul(pWeights[i])).Quo(graph.Precision)
			}
		}

		Δ = sdk.ZeroUint()

		for key, value := range graph.Nodes {
			var diff sdk.Uint
			// if _, ok := nodes[key]; ok == false {
			// 	nodes[key] = sdk.ZeroUint()
			// }
			if value.PRank.LT(nodes[key]) {
				diff = nodes[key].Sub(value.PRank)
			} else {
				diff = value.PRank.Sub(nodes[key])
			}
			Δ = Δ.Add(diff)
		}
		iter++
	}

	// fmt.Println("iterations:", iter, "Δ", Δ)
	graph.processResults(callback)
}

// make sure the total start sum of all scores is 1
// we initialze the start scores to optimize the computation
func (graph Graph) initScores(N sdk.Uint, pWeights []sdk.Uint) {
	// get sum of all node scores
	personalization := graph.Params.Personalization
	totalScore := sdk.ZeroUint()
	for _, node := range graph.Nodes {
		totalScore = totalScore.Add(node.PRank)
	}

	// if start sum is close to 1 we are done
	if totalScore.GT(graph.Precision.MulUint64(9).QuoUint64(10)) {
		return
	}

	// TODO use prev scores for initialization
	if len(pWeights) == 0 {
		// initialize all nodes if there is no personalizeation vector
		for key := range graph.Nodes {
			graph.Nodes[key].PRank = graph.Nodes[key].PRank.Add(graph.Precision.Sub(totalScore).Quo(N))
		}
		return
	}
	// initialize personalization vector
	for i, root := range personalization {
		graph.Nodes[root].PRank = graph.Nodes[root].PRank.Add(graph.Precision.Sub(totalScore).Mul(graph.Precision).Quo(pWeights[i]))
	}
}

// compute personalization weights based on degree
// this ensures source nodes will have the same weight
// we also update start scores here
func (graph Graph) initPersonalizationNodes() []sdk.Uint {
	pVector := graph.Params.Personalization
	pWeights := make([]sdk.Uint, len(pVector))

	pWeightsSum := sdk.ZeroUint()
	scoreSum := sdk.ZeroUint()
	for i, key := range pVector {
		var d sdk.Uint
		// root node score and weight should not be 0
		if d = graph.Precision; graph.Nodes[key].degree.GT(sdk.ZeroUint()) {
			d = graph.Nodes[key].degree
		}
		pWeights[i] = d
		pWeightsSum = pWeightsSum.Add(d)
		scoreSum = scoreSum.Add(graph.Nodes[key].PRank)
	}

	// normalize personalization weights
	for i, key := range pVector {
		pWeights[i] = pWeights[i].Mul(graph.Precision).Quo(pWeightsSum)
		graph.Nodes[key].PRank = scoreSum.Mul(pWeights[i]).Quo(graph.Precision)
	}

	return pWeights
}
