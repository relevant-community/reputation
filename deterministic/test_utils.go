package reputation_det

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FtoBD converts a float64 to a Uint with decimals (i * 10^decimals)
func FtoBD(n float64) sdk.Uint {
	return sdk.NewUint(uint64(n * math.Pow(10, float64(Decimals))))
}

func NewGraphHelper(α, ε float64, negConsumerRank sdk.Uint) *Graph {
	return NewGraph(FtoBD(α), FtoBD(ε), negConsumerRank)
}

func NewNodeInputHelper(id string, pRank float64, nRank float64) NodeInput {
	return NewNodeInput(id, FtoBD(pRank), FtoBD(nRank))
}

func (graph Graph) LinkHelper(source, target NodeInput, weight float64) {
	weightInt := sdk.NewInt(int64(weight * math.Pow(10, float64(Decimals/2))))
	weightInt = weightInt.Mul(sdk.NewInt(int64(math.Pow(10, float64(Decimals/2)))))
	graph.Link(source, target, weightInt)
}
