# Relevant Reputation Protocol

deterministic version: [![GoDoc](https://godoc.org/github.com/relevant-community/reputation/deterministic?status.svg)](https://godoc.org/github.com/relevant-community/reputation/deterministic) [![GoCover](http://gocover.io/_badge/github.com/relevant-community/reputation/deterministic?service=github)](http://gocover.io/github.com/relevant-community/reputation/deterministic) [![Go Report Card](https://goreportcard.com/badge/github.com/relevant-community/reputation/deterministic)](https://goreportcard.com/report/github.com/relevant-community/reputation/deterministic)

non-deterministic version: [![GoDoc](https://godoc.org/github.com/relevant-community/reputation/non-deterministic?status.svg)](https://godoc.org/github.com/relevant-community/reputation/non-deterministic) [![GoCover](http://gocover.io/_badge/github.com/relevant-community/reputation/non-deterministic?service=github)](http://gocover.io/github.com/relevant-community/reputation/non-deterministic) [![Go Report Card](https://goreportcard.com/badge/github.com/relevant-community/reputation/non-deterministic)](https://goreportcard.com/report/github.com/relevant-community/reputation/non-deterministic)

This is a personalized pagerank algorithm that supports negative links.
It can be used for voting, governance and ranking of data

Personalized version offers sybil resistant properties that can be used in a decentralized environment.

loosely based on this pagerank implementation:
https://github.com/alixaxel/pagerank

TODO: Deterministic version
TODO: Optimization - using int-indexed maps?
TODO: edge case (only impacts display) - if a node has no inputs we should set its score to 0 to avoid a stale score if all of nodes inputs are cancelled out would need to keep track of node inputs...
