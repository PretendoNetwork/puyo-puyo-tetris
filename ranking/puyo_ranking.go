package ranking

import (
	nextypes "github.com/PretendoNetwork/nex-go/v2/types"
	"github.com/PretendoNetwork/nex-protocols-go/v2/ranking/types"
)

// PROBLEM: We don't get PrincipalID here. Puyo actually uses that if it wants the ranking of a particular user (like your own rank)
func GetRankingsAndCountByCategoryAndRankingOrderParam(category *nextypes.PrimitiveU32, rankingOrderParam *types.RankingOrderParam) (*nextypes.List[*types.RankingRankData], uint32, error) {
	return nil, 0, nil
}
