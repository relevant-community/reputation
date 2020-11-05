package reputation_det

import (
	"reflect"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Result struct {
	pRank sdk.Uint
	nRank sdk.Uint
}

var zero = sdk.ZeroUint()

func TestEmpty(t *testing.T) {
	graph := NewGraphHelper(0.85, 0.000001, zero)

	actual := map[string]Result{}
	expected := map[string]Result{}

	graph.Rank(func(id string, pRank sdk.Uint, nRank sdk.Uint) {
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
	graph := NewGraphHelper(0.85, 0.000001, zero)

	a := NewNodeInputHelper("a", 0, 0)
	b := NewNodeInputHelper("b", 0, 0)
	c := NewNodeInputHelper("c", 0, 0)
	d := NewNodeInputHelper("d", 0, 0)

	// circle
	graph.LinkHelper(a, b, 1.0)
	graph.LinkHelper(b, c, 1.0)
	graph.LinkHelper(c, d, 1.0)
	graph.LinkHelper(d, a, 1.0)

	actual := map[string]Result{}
	expected := map[string]Result{
		"a": {pRank: FtoBD(0.25), nRank: FtoBD(0)},
		"b": {pRank: FtoBD(0.25), nRank: FtoBD(0)},
		"c": {pRank: FtoBD(0.25), nRank: FtoBD(0)},
		"d": {pRank: FtoBD(0.25), nRank: FtoBD(0)},
	}

	graph.Rank(func(id string, pRank sdk.Uint, nRank sdk.Uint) {
		actual[id] = Result{
			pRank: pRank,
			nRank: nRank,
		}
	})

	for key, value := range actual {
		expVal := expected[key]
		if !value.pRank.Equal(expVal.pRank) {
			t.Error("Expected", expVal.pRank.String(), "but got", value.pRank.String())
		}
		if !value.nRank.Equal(expVal.nRank) {
			t.Error("Expected", expVal.nRank.String(), "but got", value.nRank.String())
		}
	}
}

func TestWeighted(t *testing.T) {
	graph := NewGraphHelper(0.85, 0.000001, zero)

	a := NewNodeInputHelper("a", 0, 0)
	b := NewNodeInputHelper("b", 0, 0)
	c := NewNodeInputHelper("c", 0, 0)

	graph.LinkHelper(a, b, 1.0)
	graph.LinkHelper(a, c, 2.0)

	actual := map[string]Result{}

	graph.Rank(func(id string, pRank sdk.Uint, nRank sdk.Uint) {
		actual[id] = Result{
			pRank: pRank,
			nRank: nRank,
		}
	})

	if actual["b"].pRank.GTE(actual["c"].pRank) {
		t.Errorf("rank of b %s is not > c %s", actual["b"].pRank.String(), actual["c"].pRank.String())
	}
}

func TestPersonalized(t *testing.T) {
	graph := NewGraphHelper(0.85, 0.000001, zero)

	a := NewNodeInputHelper("a", 0, 0)
	b := NewNodeInputHelper("b", 0, 0)
	c := NewNodeInputHelper("c", 0, 0)
	d := NewNodeInputHelper("d", 0, 0)

	graph.AddPersonalizationNode(a)

	graph.LinkHelper(a, b, 1.0)
	graph.LinkHelper(d, c, 2.0)

	actual := map[string]Result{}
	expected := map[string]Result{
		"a": {pRank: FtoBD(0.25), nRank: FtoBD(0)},
		"b": {pRank: FtoBD(0.25), nRank: FtoBD(0)},
		"c": {pRank: FtoBD(0), nRank: FtoBD(0)},
		"d": {pRank: FtoBD(0), nRank: FtoBD(0)},
	}

	graph.Rank(func(id string, pRank sdk.Uint, nRank sdk.Uint) {
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

func TestPersonalizedNoLinkHelper(t *testing.T) {
	graph := NewGraphHelper(0.85, 0.000001, zero)

	a := NewNodeInputHelper("a", 0, 0)
	b := NewNodeInputHelper("b", 0, 0)
	c := NewNodeInputHelper("c", 0, 0)
	d := NewNodeInputHelper("d", 0, 0)

	graph.AddPersonalizationNode(a)

	graph.LinkHelper(b, c, 1.0)
	graph.LinkHelper(d, c, 1.0)

	actual := map[string]Result{}

	expected := map[string]Result{
		"a": {pRank: FtoBD(1.0), nRank: FtoBD(0)},
		"b": {pRank: FtoBD(0), nRank: FtoBD(0)},
		"c": {pRank: FtoBD(0), nRank: FtoBD(0)},
		"d": {pRank: FtoBD(0), nRank: FtoBD(0)},
	}

	graph.Rank(func(id string, pRank sdk.Uint, nRank sdk.Uint) {
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
	graph := NewGraphHelper(0.85, 0.000001, zero)

	a := NewNodeInputHelper("a", 0, 0)
	b := NewNodeInputHelper("b", 0, 0)
	c := NewNodeInputHelper("c", 0, 0)
	d := NewNodeInputHelper("d", 0, 0)

	graph.AddPersonalizationNode(a)

	graph.LinkHelper(a, b, 1.0)
	graph.LinkHelper(a, b, -1.0)
	graph.LinkHelper(a, c, 2.0)
	graph.LinkHelper(a, c, -1.0)
	graph.LinkHelper(a, d, 1.0)
	graph.LinkHelper(a, d, -2.0)

	actual := map[string]Result{}

	graph.Rank(func(id string, pRank sdk.Uint, nRank sdk.Uint) {
		actual[id] = Result{
			pRank: pRank,
			nRank: nRank,
		}
	})

	if !actual["b"].pRank.Add(actual["b"].nRank).Equal(sdk.ZeroUint()) {
		t.Errorf("rank of b should be 0 but its %s", actual["b"].pRank.String())
	}

	if !actual["c"].nRank.Equal(sdk.ZeroUint()) {
		t.Errorf("c rank should be positive")
	}

	if !actual["d"].pRank.Equal(sdk.ZeroUint()) {
		t.Errorf("d rank should be negative")
	}
}

func TestNegativeLink(t *testing.T) {
	graph := NewGraphHelper(0.85, 0.000001, zero)

	a := NewNodeInputHelper("a", 0, 0)
	b := NewNodeInputHelper("b", 0, 0)
	c := NewNodeInputHelper("c", 0, 0)
	d := NewNodeInputHelper("d", 0, 0)
	e := NewNodeInputHelper("e", 0, 0)

	graph.AddPersonalizationNode(a)

	graph.LinkHelper(a, b, 2.0)
	graph.LinkHelper(a, c, 1.0)
	graph.LinkHelper(c, d, 1.0)
	graph.LinkHelper(b, d, -1.0)
	graph.LinkHelper(a, e, -1.0)

	actual := map[string]Result{}

	graph.Rank(func(id string, pRank sdk.Uint, nRank sdk.Uint) {
		actual[id] = Result{
			pRank: pRank,
			nRank: nRank,
		}
	})

	if actual["d"].nRank.Sub(actual["d"].pRank).LTE(sdk.ZeroUint()) {
		t.Errorf("rank of d should be neagative")
	}

	if !actual["e"].pRank.Equal(sdk.ZeroUint()) || actual["e"].nRank.Equal(sdk.ZeroUint()) {
		t.Errorf("pure neagative node has incorrect results")
	}

	// use prev computation as input for the next iteration
	initNegativeConsumer := actual["negConsumer"].pRank
	if _, ok := actual["negConsumer"]; ok == false {
		initNegativeConsumer = zero
	}

	graph = NewGraphHelper(0.85, 0.000001, initNegativeConsumer)

	a = NewNodeInput("a", actual["a"].pRank, actual["a"].nRank)
	b = NewNodeInput("b", actual["b"].pRank, actual["b"].nRank)
	c = NewNodeInput("c", actual["c"].pRank, actual["c"].nRank)
	d = NewNodeInput("d", actual["d"].pRank, actual["d"].nRank)
	e = NewNodeInput("e", actual["e"].pRank, actual["e"].nRank)

	graph.AddPersonalizationNode(a)

	graph.LinkHelper(a, b, 2.0)
	graph.LinkHelper(a, c, 1.0)
	graph.LinkHelper(c, d, 1.0)
	graph.LinkHelper(b, d, -1.0)
	graph.LinkHelper(d, e, 1.0)

	graph.Rank(func(id string, pRank sdk.Uint, nRank sdk.Uint) {
		actual[id] = Result{
			pRank: pRank,
			nRank: nRank,
		}
	})

	if !actual["e"].pRank.Equal(sdk.ZeroUint()) {
		t.Errorf("weight of neg node should be 0")
	}
}

func TestNegativeConsumer(t *testing.T) {
	graph := NewGraphHelper(0.85, 0.000001, zero)

	a := NewNodeInputHelper("a", 0, 0)
	b := NewNodeInputHelper("b", 0, 0)
	c := NewNodeInputHelper("c", 0, 0)
	d := NewNodeInputHelper("d", 0, 0)
	e := NewNodeInputHelper("e", 0, 0)

	graph.AddPersonalizationNode(a)

	graph.LinkHelper(a, b, 1.0)
	graph.LinkHelper(a, c, 2.0)
	graph.LinkHelper(c, d, 1.0)
	graph.LinkHelper(b, d, -1.0)
	graph.LinkHelper(d, e, 1.0)

	actual := map[string]Result{}

	graph.Rank(func(id string, pRank sdk.Uint, nRank sdk.Uint) {
		actual[id] = Result{
			pRank: pRank,
			nRank: nRank,
		}
	})

	// use prev computation as input for the next iteration
	initNegativeConsumer := actual["negConsumer"].pRank
	if _, ok := actual["negConsumer"]; ok == false {
		initNegativeConsumer = zero
	}

	graph = NewGraphHelper(0.85, 0.000001, initNegativeConsumer)

	eRank := actual["e"].pRank

	a = NewNodeInput("a", actual["a"].pRank, actual["a"].nRank)
	b = NewNodeInput("b", actual["b"].pRank, actual["b"].nRank)
	c = NewNodeInput("c", actual["c"].pRank, actual["c"].nRank)
	d = NewNodeInput("d", actual["d"].pRank, actual["d"].nRank)
	e = NewNodeInput("e", actual["e"].pRank, actual["e"].nRank)

	graph.AddPersonalizationNode(a)

	graph.LinkHelper(a, b, 1.0)
	graph.LinkHelper(a, c, 2.0)
	graph.LinkHelper(c, d, 1.0)
	graph.LinkHelper(b, d, -1.0)
	graph.LinkHelper(d, e, 1.0)

	graph.Rank(func(id string, pRank sdk.Uint, nRank sdk.Uint) {
		actual[id] = Result{
			pRank: pRank,
			nRank: nRank,
		}
	})

	if eRank.LT(actual["e"].pRank) {
		t.Errorf("weight of neg node should decrease")
	}
}

func TestMaxNeg(t *testing.T) {
	graph := NewGraphHelper(0.85, 0.000001, zero)

	a := NewNodeInputHelper("a", 0, 0)
	b := NewNodeInputHelper("b", 0, 0)
	c := NewNodeInputHelper("c", 0, 0)
	d := NewNodeInputHelper("d", 0, 0)
	e := NewNodeInputHelper("e", 0, 0)

	graph.AddPersonalizationNode(a)

	graph.LinkHelper(a, b, 11.0)
	graph.LinkHelper(a, c, 12.0)
	graph.LinkHelper(c, d, 1.0)
	graph.LinkHelper(b, d, -1.0)
	graph.LinkHelper(d, e, 1.0)

	actual := map[string]Result{}

	graph.Rank(func(id string, pRank sdk.Uint, nRank sdk.Uint) {
		actual[id] = Result{
			pRank: pRank,
			nRank: nRank,
		}
	})

	// use prev computation as input for the next iteration
	initNegativeConsumer := actual["negConsumer"].pRank
	if _, ok := actual["negConsumer"]; ok == false {
		initNegativeConsumer = zero
	}

	graph = NewGraphHelper(0.85, 0.000001, initNegativeConsumer)

	eRank := actual["e"].pRank

	a = NewNodeInput("a", actual["a"].pRank, actual["a"].nRank)
	b = NewNodeInput("b", actual["b"].pRank, actual["b"].nRank)
	c = NewNodeInput("c", actual["c"].pRank, actual["c"].nRank)
	d = NewNodeInput("d", actual["d"].pRank, actual["d"].nRank)
	e = NewNodeInput("e", actual["e"].pRank, actual["e"].nRank)

	graph.AddPersonalizationNode(a)

	graph.LinkHelper(a, b, 11.0)
	graph.LinkHelper(a, c, 12.0)
	graph.LinkHelper(c, d, 1.0)
	graph.LinkHelper(b, d, -1.0)
	graph.LinkHelper(d, e, 1.0)

	graph.Rank(func(id string, pRank sdk.Uint, nRank sdk.Uint) {
		actual[id] = Result{
			pRank: pRank,
			nRank: nRank,
		}
	})

	if eRank.LT(actual["e"].pRank) {
		t.Errorf("weight of neg node should decrease")
	}
}
