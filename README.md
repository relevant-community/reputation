# Relevant Reputation Protocol [![Go Report Card](https://goreportcard.com/badge/github.com/relevant-community/reputation)](https://goreportcard.com/report/github.com/relevant-community/reputation)

deterministic version: [![GoDoc](https://godoc.org/github.com/relevant-community/reputation/deterministic?status.svg)](https://godoc.org/github.com/relevant-community/reputation/deterministic) [![GoCover](http://gocover.io/_badge/github.com/relevant-community/reputation/deterministic)](http://gocover.io/github.com/relevant-community/reputation/deterministic)

non-deterministic version: [![GoDoc](https://godoc.org/github.com/relevant-community/reputation/non-deterministic?status.svg)](https://godoc.org/github.com/relevant-community/reputation/non-deterministic) [![GoCover](http://gocover.io/_badge/github.com/relevant-community/reputation/non-deterministic)](http://gocover.io/github.com/relevant-community/reputation/non-deterministic)

## Abstract

Relevant Reputation Protocol is a personalized pagerank algorithm that supports negative links. It is used to compute user and content rankings in the [Relevant App](https://relevant.community), a reddit-like bulletin board.

## Use Cases

Because the algorithm supports negative links, it can be used to represent upvotes and downvotes. This enables usescases such as voting, governance and ranking of data.

## Deterministic Version

The deterministiv version of the algorithm uses `Uint` and safe-math libs from [Cosmos Sdk](https://github.com/cosmos/cosmos-sdk) to avoid floating point computations.

The deterministic version can be used in an environment where concensus is required.

## Usage

create an instance of the graph

```go
graph := rep.NewGraph(0.85, 1e-8, 0)
```

explanation of the graph params:

1.  `α` is the probability of not doing a jump during a random walk, usually `.85`
2.  `ε` is the error margin required to approximate convergence, usually something small
3.  The third parameter is a cached rank of the negative consumer node. This is an extra node we add to the graph to enable negative links, and this is its rank from previous computations.

## Core Concepts and Features

### Personalization

The `personalization` property of the graph designates trusted/authority nodes. These nodes will have a high reputation by default. Repuation flows from these nodes to the rest of the graph.

Personalized pagerank allows for construction of a sybil-resistant network. This means users that are not part of the personalization vector cannot manipulate the scores by creating sybil nodes that link to (or upvote) one another.

Notes:

- If the 'personalization' vector is left empty, the algorithm will not be sybil resistant and malicious nodes will be able to manipulate the rankings.
- Nodes in the personalization vector have the power to manipulate rankings, so special care should be taken to when selecting these nodes.

### Negative Links

Negative links are possible because we represent each entity in our graph via two nodes - one positive and one negative. Negative links boost the ranking of the negative node, positive links boost the ranking of the positive node. These two scores can then be merged into a single reputation score.

However we cannot easily use the merged score when creating links. When a node has both a positive and a negative score, only the positive node is used for outgoing links, but we modulate the weight of the outgoing linke based on the negative score.

This is done by running the pagerank computation twice. The first round ensures we have computed negative and positive rankings of the nodes. The second round enables us to take the negative score into account and ignore nodes that have a high negative/positive rank ratio all together.

**Implementation details:**
We modulate the weight of outgoing links by creating one global `negConsumer` node. Nodes that have both a negative and a positive rank, will have a portion of their outgoing weight consumed by a link to `negConsumer`, thereby decreasing the weight of other outgoing links.

## TODOS:

- [ ] Optimization & benchmarking - we should be using int-indexed maps and fixed size arrays where possible.

- [ ] Edge case (only impacts display) - if a node has no inputs we should re-set its score to 0 to avoid a stale score being displayed after all links to the node were removed or cancelled-out.

## Credits

This pagerank implementation was used as a starting point.
https://github.com/alixaxel/pagerank
