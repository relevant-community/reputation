package rep

import (
	"reflect"
	"testing"
)

type Result struct {
	pRank float64
	nRank float64
}

func TestEmpty(t *testing.T) {
	graph := NewGraph(0.85, 0.000001, 0)

	actual := map[string]Result{}
	expected := map[string]Result{}

	graph.Rank(func(id string, pRank float64, nRank float64) {
		actual[id] = Result{
			pRank: pRank,
			nRank: nRank,
		}
	})

	if reflect.DeepEqual(actual, expected) != true {
		t.Error("Expected", expected, "but got", actual)
	}
}

func TestSimple(t *testing.T) {
	graph := NewGraph(0.85, 0.000001, 0)

	a := NewNodeInput("a", 0, 0)
	b := NewNodeInput("b", 0, 0)
	c := NewNodeInput("c", 0, 0)
	d := NewNodeInput("d", 0, 0)

	// circle
	graph.Link(a, b, 1.0)
	graph.Link(b, c, 1.0)
	graph.Link(c, d, 1.0)
	graph.Link(d, a, 1.0)

	actual := map[string]Result{}
	expected := map[string]Result{
		"a": {pRank: 0.25, nRank: 0},
		"b": {pRank: 0.25, nRank: 0},
		"c": {pRank: 0.25, nRank: 0},
		"d": {pRank: 0.25, nRank: 0},
	}

	graph.Rank(func(id string, pRank float64, nRank float64) {
		actual[id] = Result{
			pRank: pRank,
			nRank: nRank,
		}
	})

	if reflect.DeepEqual(actual, expected) != true {
		t.Error("Expected", expected, "but got", actual)
	}
}

func TestWeighted(t *testing.T) {
	graph := NewGraph(0.85, 0.000001, 0)

	a := NewNodeInput("a", 0, 0)
	b := NewNodeInput("b", 0, 0)
	c := NewNodeInput("c", 0, 0)

	graph.Link(a, b, 1.0)
	graph.Link(a, c, 2.0)

	actual := map[string]Result{}

	graph.Rank(func(id string, pRank float64, nRank float64) {
		actual[id] = Result{
			pRank: pRank,
			nRank: nRank,
		}
	})

	if actual["b"].pRank >= actual["c"].pRank {
		t.Errorf("rank of b %f is not > c %f", actual["b"].pRank, actual["c"].pRank)
	}
}

func TestPersonalized(t *testing.T) {
	graph := NewGraph(0.85, 0.000001, 0)

	a := NewNodeInput("a", 0, 0)
	b := NewNodeInput("b", 0, 0)
	c := NewNodeInput("c", 0, 0)
	d := NewNodeInput("d", 0, 0)

	graph.AddPersonalizationNode(a)

	graph.Link(a, b, 1.0)
	graph.Link(d, c, 2.0)

	actual := map[string]Result{}
	expected := map[string]Result{
		"a": {pRank: 0.25, nRank: 0},
		"b": {pRank: 0.25, nRank: 0},
		"c": {pRank: 0, nRank: 0},
		"d": {pRank: 0, nRank: 0},
	}

	graph.Rank(func(id string, pRank float64, nRank float64) {
		actual[id] = Result{
			pRank: pRank,
			nRank: nRank,
		}
	})

	if reflect.DeepEqual(actual["c"], expected["c"]) != true {
		t.Error("Expected", expected["c"], "but got", actual["c"])
	}
	if reflect.DeepEqual(actual["d"], expected["d"]) != true {
		t.Error("Expected", expected["d"], "but got", actual["d"])
	}
}

func TestPersonalizedNoLink(t *testing.T) {
	graph := NewGraph(0.85, 0.000001, 0)

	a := NewNodeInput("a", 0, 0)
	b := NewNodeInput("b", 0, 0)
	c := NewNodeInput("c", 0, 0)
	d := NewNodeInput("d", 0, 0)

	graph.AddPersonalizationNode(a)

	graph.Link(b, c, 1.0)
	graph.Link(d, c, 1.0)

	actual := map[string]Result{}

	expected := map[string]Result{
		"a": {pRank: 1.0, nRank: 0},
		"b": {pRank: 0, nRank: 0},
		"c": {pRank: 0, nRank: 0},
		"d": {pRank: 0, nRank: 0},
	}

	graph.Rank(func(id string, pRank float64, nRank float64) {
		actual[id] = Result{
			pRank: pRank,
			nRank: nRank,
		}
	})

	if reflect.DeepEqual(actual, expected) != true {
		t.Error("Expected", expected, "but got", actual)
	}
}

func TestCancelOpposites(t *testing.T) {
	graph := NewGraph(0.85, 0.000001, 0)

	a := NewNodeInput("a", 0, 0)
	b := NewNodeInput("b", 0, 0)
	c := NewNodeInput("c", 0, 0)
	d := NewNodeInput("d", 0, 0)

	graph.AddPersonalizationNode(a)

	graph.Link(a, b, 1.0)
	graph.Link(a, b, -1.0)
	graph.Link(a, c, 2.0)
	graph.Link(a, c, -1.0)
	graph.Link(a, d, 1.0)
	graph.Link(a, d, -2.0)

	actual := map[string]Result{}

	graph.Rank(func(id string, pRank float64, nRank float64) {
		actual[id] = Result{
			pRank: pRank,
			nRank: nRank,
		}
	})

	if actual["b"].pRank+actual["b"].nRank != 0 {
		t.Errorf("rank of b should be 0")
	}

	if actual["c"].nRank != 0 {
		t.Errorf("c rank should be positive")
	}

	if actual["d"].pRank != 0 {
		t.Errorf("d rank should be negative")
	}
}

func TestNegativeLink(t *testing.T) {
	graph := NewGraph(0.85, 0.000001, 0)

	a := NewNodeInput("a", 0, 0)
	b := NewNodeInput("b", 0, 0)
	c := NewNodeInput("c", 0, 0)
	d := NewNodeInput("d", 0, 0)
	e := NewNodeInput("e", 0, 0)

	graph.AddPersonalizationNode(a)

	graph.Link(a, b, 2.0)
	graph.Link(a, c, 1.0)
	graph.Link(c, d, 1.0)
	graph.Link(b, d, -1.0)
	graph.Link(a, e, -1.0)

	actual := map[string]Result{}

	graph.Rank(func(id string, pRank float64, nRank float64) {
		actual[id] = Result{
			pRank: pRank,
			nRank: nRank,
		}
	})

	if actual["d"].pRank-actual["d"].nRank >= 0 {
		t.Errorf("rank of d should be neagative")
	}

	if actual["e"].pRank != 0 || actual["e"].nRank == 0 {
		t.Errorf("pure neagative node has incorrect results")
	}

	// use prev computation as input for the next iteration

	graph = NewGraph(0.85, 0.000001, actual["negConsumer"].pRank)

	a = NewNodeInput("a", actual["a"].pRank, actual["a"].nRank)
	b = NewNodeInput("b", actual["b"].pRank, actual["b"].nRank)
	c = NewNodeInput("c", actual["c"].pRank, actual["c"].nRank)
	d = NewNodeInput("d", actual["d"].pRank, actual["d"].nRank)
	e = NewNodeInput("e", actual["e"].pRank, actual["e"].nRank)

	graph.AddPersonalizationNode(a)

	graph.Link(a, b, 2.0)
	graph.Link(a, c, 1.0)
	graph.Link(c, d, 1.0)
	graph.Link(b, d, -1.0)
	graph.Link(d, e, 1.0)

	graph.Rank(func(id string, pRank float64, nRank float64) {
		actual[id] = Result{
			pRank: pRank,
			nRank: nRank,
		}
	})

	if actual["e"].pRank != 0 {
		t.Errorf("weight of neg node should be 0")
	}
}

func TestNegativeConsumer(t *testing.T) {
	graph := NewGraph(0.85, 0.000001, 0)

	a := NewNodeInput("a", 0, 0)
	b := NewNodeInput("b", 0, 0)
	c := NewNodeInput("c", 0, 0)
	d := NewNodeInput("d", 0, 0)
	e := NewNodeInput("e", 0, 0)

	graph.AddPersonalizationNode(a)

	graph.Link(a, b, 1.0)
	graph.Link(a, c, 2.0)
	graph.Link(c, d, 1.0)
	graph.Link(b, d, -1.0)
	graph.Link(d, e, 1.0)

	actual := map[string]Result{}

	graph.Rank(func(id string, pRank float64, nRank float64) {
		actual[id] = Result{
			pRank: pRank,
			nRank: nRank,
		}
	})

	// use prev computation as input for the next iteration

	graph = NewGraph(0.85, 0.000001, actual["negConsumer"].pRank)

	eRank := actual["e"].pRank

	a = NewNodeInput("a", actual["a"].pRank, actual["a"].nRank)
	b = NewNodeInput("b", actual["b"].pRank, actual["b"].nRank)
	c = NewNodeInput("c", actual["c"].pRank, actual["c"].nRank)
	d = NewNodeInput("d", actual["d"].pRank, actual["d"].nRank)
	e = NewNodeInput("e", actual["e"].pRank, actual["e"].nRank)

	graph.AddPersonalizationNode(a)

	graph.Link(a, b, 1.0)
	graph.Link(a, c, 2.0)
	graph.Link(c, d, 1.0)
	graph.Link(b, d, -1.0)
	graph.Link(d, e, 1.0)

	graph.Rank(func(id string, pRank float64, nRank float64) {
		actual[id] = Result{
			pRank: pRank,
			nRank: nRank,
		}
	})

	if eRank <= actual["e"].pRank {
		t.Errorf("weight of neg node should decrease %f, %f", eRank, actual["e"].pRank)
	}
}

func TestMaxNeg(t *testing.T) {
	graph := NewGraph(0.85, 0.000001, 0)

	a := NewNodeInput("a", 0, 0)
	b := NewNodeInput("b", 0, 0)
	c := NewNodeInput("c", 0, 0)
	d := NewNodeInput("d", 0, 0)
	e := NewNodeInput("e", 0, 0)

	graph.AddPersonalizationNode(a)

	graph.Link(a, b, MAX_NEG_OFFSET+1)
	graph.Link(a, c, MAX_NEG_OFFSET+2)
	graph.Link(c, d, 1.0)
	graph.Link(b, d, -1.0)
	graph.Link(d, e, 1.0)

	actual := map[string]Result{}

	graph.Rank(func(id string, pRank float64, nRank float64) {
		actual[id] = Result{
			pRank: pRank,
			nRank: nRank,
		}
	})

	// use prev computation as input for the next iteration

	graph = NewGraph(0.85, 0.000001, actual["negConsumer"].pRank)

	eRank := actual["e"].pRank

	a = NewNodeInput("a", actual["a"].pRank, actual["a"].nRank)
	b = NewNodeInput("b", actual["b"].pRank, actual["b"].nRank)
	c = NewNodeInput("c", actual["c"].pRank, actual["c"].nRank)
	d = NewNodeInput("d", actual["d"].pRank, actual["d"].nRank)
	e = NewNodeInput("e", 0, 0)

	graph.AddPersonalizationNode(a)

	graph.Link(a, b, MAX_NEG_OFFSET+1)
	graph.Link(a, c, MAX_NEG_OFFSET+2)
	graph.Link(c, d, 1.0)
	graph.Link(b, d, -1.0)
	graph.Link(d, e, 1.0)

	actual = map[string]Result{}

	graph.Rank(func(id string, pRank float64, nRank float64) {
		actual[id] = Result{
			pRank: pRank,
			nRank: nRank,
		}
	})

	if eRank <= actual["e"].pRank {
		t.Errorf("weight of neg node should decrease %f, %f", eRank, actual["e"].pRank)
	}
}
