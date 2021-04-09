package main

import (
	"math"
	"strings"

	"github.com/go-echarts/go-echarts/v2/opts"
	rep "github.com/relevant-community/reputation/rep"
)

type Data struct {
	Nodes []opts.GraphNode
	Links []opts.GraphLink
}

type Result struct {
	pRank float64
	nRank float64
}

func GetGraph1() Data {
	graph := rep.NewGraph(1, 0.000001, 0)

	a := rep.NewNode("a", 0, 0)
	b := rep.NewNode("b", 0, 0)
	c := rep.NewNode("c", 0, 0)
	d := rep.NewNode("d", 0, 0)
	e := rep.NewNode("e", 0, 0)
	f := rep.NewNode("f", 0, 0)

	graph.AddPersonalizationNode(a)

	genLinks := func(graph *rep.Graph) {
		graph.Link(a, b, 1.0)
		graph.Link(a, c, 2.0)
		graph.Link(a, f, -1.0)

		graph.Link(c, d, 1.0)
		graph.Link(b, d, -1.0)
		graph.Link(d, e, 1.0)
		graph.Link(f, e, 2.0)

	}

	genLinks(graph)
	firstRound := map[string]Result{}

	graph.Rank(func(id string, pRank float64, nRank float64) {
		firstRound[id] = Result{
			pRank: pRank,
			nRank: nRank,
		}
	})

	// use prev computation as input for the next iteration

	graph = rep.NewGraph(1, 0.000001, firstRound["negConsumer"].pRank)

	a = rep.NewNode("a", firstRound["a"].pRank, firstRound["a"].nRank)
	b = rep.NewNode("b", firstRound["b"].pRank, firstRound["b"].nRank)
	c = rep.NewNode("c", firstRound["c"].pRank, firstRound["c"].nRank)
	d = rep.NewNode("d", firstRound["d"].pRank, firstRound["d"].nRank)
	e = rep.NewNode("e", firstRound["e"].pRank, firstRound["e"].nRank)
	f = rep.NewNode("f", firstRound["f"].pRank, firstRound["f"].nRank)

	graph.AddPersonalizationNode(a)

	genLinks(graph)

	var links []opts.GraphLink

	var nodes []opts.GraphNode
	graph.Rank(func(id string, pRank float64, nRank float64) {
		rank := pRank - nRank

		var symbol string
		var category int
		switch true {
		case id == "negConsumer":
			symbol = "diamond"
			category = 3
		case contains(graph.Params.Personalization, id):
			symbol = "triangle"
			category = 0
		default:
			symbol = "circle"
			category = 2
		}

		color := "#39FF14"
		if rank < 0 {
			color = "pink"
			category = 1
		}

		nodes = append(nodes, opts.GraphNode{
			Name:       id,
			Value:      float32(rank),
			Symbol:     symbol,
			SymbolSize: math.Abs(rank) * 350,
			Category:   category,
			ItemStyle: &opts.ItemStyle{
				Color: color,
			},
		})
	})

	for source, edges := range graph.Edges {
		for target, edge := range edges {
			value := float32(edge)
			if strings.Contains(target, "_1") {
				value = -float32(edge)
			}
			links = append(links, opts.GraphLink{
				Source: source,
				Target: strings.Replace(target, "_1", "", -1),
				Value:  value,
			})
		}
	}

	return Data{
		Nodes: nodes,
		Links: links,
	}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
