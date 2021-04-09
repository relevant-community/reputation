package main

import (
	"io"
	"net/http"
	"os"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func repGraph(data Data, title string) *charts.Graph {
	graph := charts.NewGraph()
	graph.SetGlobalOptions(
		// charts.WithLegendOpts(opts.Legend{
		// 	Show: true,
		// }),
		charts.WithTitleOpts(opts.Title{
			Title: title,
		}))

	graph.AddSeries("graph", data.Nodes, data.Links).
		SetSeriesOptions(
			charts.WithGraphChartOpts(opts.GraphChart{
				Categories: []*opts.GraphCategory{
					{Name: "Personalization"},
					{Name: "Negative"},
					{Name: "Positive"},
					{Name: "Neg Consumer"},
				},
				Force:              &opts.GraphForce{Repulsion: 2000},
				Layout:             "force",
				Roam:               true,
				FocusNodeAdjacency: true,
			}),
			charts.WithLabelOpts(opts.Label{Show: true, Position: "right", Color: "black"}),

			charts.WithEmphasisOpts(opts.Emphasis{
				Label: &opts.Label{
					Formatter: "rank: {c}",
					Show:      true,
					Color:     "black",
				},
			}),
			// charts.WithLineStyleOpts(opts.LineStyle{
			// 	Curveness: 0.3,
			// }),
		)
	return graph
}

func renderGraph() {
	page := components.NewPage()
	data1 := GetGraph1()
	title1 := `
		a: personalization node
		d: node with negative and positive rank
		f: node with negative rank
		`

	page.AddCharts(
		repGraph(data1, title1),
	)

	f, err := os.Create("index.html")
	if err != nil {
		panic(err)

	}
	page.Render(io.MultiWriter(f))
}

func main() {
	renderGraph()
	http.Handle("/", http.FileServer(http.Dir("./")))
	http.ListenAndServe(":7000", nil)
}
