package rep

import (
	"math"
)

// Rank computes the PageRank of every node in the directed graph.
// α (alpha) is the damping factor, usually set to 0.85.
// ε (epsilon) is the convergence criteria, usually set to a tiny value.
//
// This method will run as many iterations as needed, until the graph converges.
func (graph Graph) Rank(callback func(key string, pRank float64, nRank float64)) {
	graph.Finalize()

	Δ := float64(1.0)
	N := float64(len(graph.Nodes))
	pVector := graph.Params.Personalization
	ε := graph.Params.ε
	α := graph.Params.α

	personalized := len(pVector) > 0

	// these are personlaization node weights
	// we adjust them so that all p nodes have the same outgoing link weight
	pWeights := graph.initPersonalizationNodes()

	// Normalize all the edge weights so that their sum amounts to 1.
	for source := range graph.Nodes {
		if graph.Nodes[source].degree > 0 {
			for target := range graph.Edges[source] {
				graph.Edges[source][target] /= graph.Nodes[source].degree
			}
		}
	}

	graph.initScores(N, pWeights)

	iter := 0
	for Δ > ε {
		danglingWeight := float64(0)
		nodes := map[string]float64{}

		for key, value := range graph.Nodes {
			nodes[key] = value.PRank

			if value.degree == 0 {
				danglingWeight += value.PRank
			}

			graph.Nodes[key].PRank = 0
		}

		danglingWeight *= α

		for source := range graph.Nodes {
			for target, weight := range graph.Edges[source] {
				graph.Nodes[target].PRank += α * nodes[source] * weight
			}

			if !personalized {
				graph.Nodes[source].PRank += (1-α)/N + danglingWeight/N
			}
		}

		// random jump + dangling weights are transferred to admins
		// this makes pagerank sybil resistant
		if personalized {
			for i, root := range pVector {
				graph.Nodes[root].PRank += (1 - α + danglingWeight) * pWeights[i]
			}
		}

		Δ = 0

		for key, value := range graph.Nodes {
			Δ += math.Abs(value.PRank - nodes[key])
		}
		iter++
	}

	// fmt.Println("iterations:", iter, "Δ", Δ)
	graph.processResults(callback)
}

// make sure the total start sum of all scores is 1
// we initialze the start scores to optimize the computation
func (graph Graph) initScores(N float64, pWeights []float64) {
	// get sum of all node scores
	personalization := graph.Params.Personalization
	var totalScore float64
	for _, node := range graph.Nodes {
		totalScore += node.PRank
	}

	if totalScore > .9 {
		return
	}

	// TODO use prev scores for initialization
	if len(pWeights) == 0 {
		// initialize all nodes if there is no personalizeation vector
		for key := range graph.Nodes {
			graph.Nodes[key].PRank += (1 - totalScore) / N
		}
		return
	}
	// initialize personalization vector
	for i, root := range personalization {
		graph.Nodes[root].PRank += (1 - totalScore) * pWeights[i]
	}
}

// compute personalization weights based on degree
// this ensures source nodes will have the same weight
// we also update start scores here
func (graph Graph) initPersonalizationNodes() []float64 {
	pVector := graph.Params.Personalization
	pWeights := make([]float64, len(pVector))

	var pWeightsSum float64
	var scoreSum float64
	for i, key := range pVector {
		var d float64
		// root node score and weight should not be 0
		if d = 1; graph.Nodes[key].degree > 0 {
			d = graph.Nodes[key].degree
		}
		pWeights[i] = d
		pWeightsSum += d
		scoreSum += graph.Nodes[key].PRank
	}

	// normalize personalization weights
	for i, key := range pVector {
		pWeights[i] /= pWeightsSum
		graph.Nodes[key].PRank = scoreSum * pWeights[i]
	}

	return pWeights
}
