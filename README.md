# Relevant Reputation Protocol [![Go Report Card](https://goreportcard.com/badge/github.com/relevant-community/reputation)](https://goreportcard.com/report/github.com/relevant-community/reputation)

deterministic version: [![GoDoc](https://godoc.org/github.com/relevant-community/reputation/deterministic?status.svg)](https://godoc.org/github.com/relevant-community/reputation/deterministic) [![GoCover](http://gocover.io/_badge/github.com/relevant-community/reputation/deterministic)](http://gocover.io/github.com/relevant-community/reputation/deterministic)

non-deterministic version: [![GoDoc](https://godoc.org/github.com/relevant-community/reputation/non-deterministic?status.svg)](https://godoc.org/github.com/relevant-community/reputation/non-deterministic) [![GoCover](http://gocover.io/_badge/github.com/relevant-community/reputation/non-deterministic)](http://gocover.io/github.com/relevant-community/reputation/non-deterministic)

## This is a personalized pagerank algorithm that supports negative links.

Negative links are able to represent upvotes and downvotes.
As a result, the algorithm can be used for voting, governance and ranking of data.

The deterministic version can be used in an environment where concensus is required. (Cosmos sdk Uint and Int safe-math libraries are used float-free math).

Personalized setting offers sybil resistant properties that can be used in a decentralized environment.

This pagerank implementation was used as a starting point.
https://github.com/alixaxel/pagerank

TODO: Optimization & benchmarking - we should be using int-indexed maps and fixed size arrays where possible.

TODO: Edge case (only impacts display) - if a node has no inputs we should re-set its score to 0 to avoid a stale score being displayed after all links to the node were removed or cancelled-out.

## Notes on negative link implementation

Negative links are possible because we represent each entity in our graph via two nodes - a positive and a negative. Negative links boost the negative node, positive links boost the positive node. These two scores can then be merged into a single score.

The challange is that when a node has both a positive and a negative score, the full positive score is used for outgoing links. We solve this by running the pagerank computation again using the results of the initial computation as a starting point. This enables us to decrease the weight of the outgoing links proportional to the entity's total score. We ignore nodes that have a high negative/positive rank ratio all together.
