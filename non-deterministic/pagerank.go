package reputation

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
	N := float64(len(graph.nodes))
	pVector := graph.params.personalization
	ε := graph.params.ε
	α := graph.params.α

	personalized := len(pVector) > 0

	// these are personlaization node weights
	// we adjust them so that all p nodes have the same outgoing link weight
	pWeights := graph.initPersonalizationNodes()

	// Normalize all the edge weights so that their sum amounts to 1.
	for source := range graph.nodes {
		if graph.nodes[source].degree > 0 {
			for target := range graph.edges[source] {
				graph.edges[source][target] /= graph.nodes[source].degree
			}
		}
	}

	graph.initScores(N, pWeights)

	iter := 0
	for Δ > ε {
		danglingWeight := float64(0)
		nodes := map[string]float64{}

		for key, value := range graph.nodes {
			nodes[key] = value.rank

			if value.degree == 0 {
				danglingWeight += value.rank
			}

			graph.nodes[key].rank = 0
		}

		danglingWeight *= α

		for source := range graph.nodes {
			for target, weight := range graph.edges[source] {
				graph.nodes[target].rank += α * nodes[source] * weight
			}

			if !personalized {
				graph.nodes[source].rank += (1-α)/N + danglingWeight/N
			}
		}

		// random jump + dangling weights are transferred to admins
		// this makes pagerank sybil resistant
		if personalized {
			for i, root := range pVector {
				graph.nodes[root].rank += (1 - α + danglingWeight) * pWeights[i]
			}
		}

		Δ = 0

		for key, value := range graph.nodes {
			Δ += math.Abs(value.rank - nodes[key])
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
	personalization := graph.params.personalization
	var totalScore float64
	for _, node := range graph.nodes {
		totalScore += node.rank
	}

	if totalScore > .9 {
		return
	}

	// TODO use prev scores for initialization
	if len(pWeights) == 0 {
		// initialize all nodes if there is no personalizeation vector
		for key := range graph.nodes {
			graph.nodes[key].rank += (1 - totalScore) / N
		}
		return
	}
	// initialize personalization vector
	for i, root := range personalization {
		graph.nodes[root].rank += (1 - totalScore) * pWeights[i]
	}
}

// compute personalization weights based on degree
// this ensures source nodes will have the same weight
// we also update start scores here
func (graph Graph) initPersonalizationNodes() []float64 {
	pVector := graph.params.personalization
	pWeights := make([]float64, len(pVector))

	var pWeightsSum float64
	var scoreSum float64
	for i, key := range pVector {
		var d float64
		// root node score and weight should not be 0
		if d = 1; graph.nodes[key].degree > 0 {
			d = graph.nodes[key].degree
		}
		pWeights[i] = d
		pWeightsSum += d
		scoreSum += graph.nodes[key].rank
	}

	// normalize personalization weights
	for i, key := range pVector {
		pWeights[i] /= pWeightsSum
		graph.nodes[key].rank = scoreSum * pWeights[i]
	}

	return pWeights
}
