# Relevant Reputation Protocol [![GoDoc](https://godoc.org/github.com/relevant-community/reputation?status.svg)](https://godoc.org/github.com/relevant-community/reputation) [![GoCover](http://gocover.io/_badge/github.com/relevant-community/reputation?service=github)](http://gocover.io/github.com/relevant-community/reputation) [![Go Report Card](https://goreportcard.com/badge/github.com/relevant-community/reputation)](https://goreportcard.com/report/github.com/relevant-community/reputation)

This is a personalized pagerank algorithm that supports negative links.
It can be used for voting, governance and ranking of data

Personalized version offers sybil resistant properties that can be used in a decentralized environment.

loosely based on this pagerank implementation:
https://github.com/alixaxel/pagerank

TODO: Deterministic version
TODO: Optimization - using int-indexed maps?
TODO: edge case (only impacts display) - if a node has no inputs we should set its score to 0 to avoid a stale score if all of nodes inputs are cancelled out would need to keep track of node inputs...
