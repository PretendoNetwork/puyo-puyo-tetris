package ranking

import (
	"database/sql"
	"errors"
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	rankingtypes "github.com/PretendoNetwork/nex-protocols-go/v2/ranking/types"
	"github.com/PretendoNetwork/pq-extended"
	"github.com/PretendoNetwork/puyo-puyo-tetris/globals"
)

var getNearbyFriendsRankings *sql.Stmt

func GetNearbyFriendsRankingsAndCountByCategoryAndRankingOrderParam(pid types.PID, category types.UInt32, rankingOrderParam rankingtypes.RankingOrderParam) (types.List[rankingtypes.RankingRankData], uint32, error) {
	friends := globals.GetUserFriendPIDs(uint32(pid))

	rows, err := getNearbyFriendsRankings.Query(
		category,
		rankingOrderParam.OrderCalculation == 0,
		rankingOrderParam.GroupIndex,
		rankingOrderParam.GroupNum,
		rankingOrderParam.Offset,
		rankingOrderParam.Length,
		pid,
		pqextended.Array(friends),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, nil
	} else if err != nil {
		return nil, 0, nex.NewError(nex.ResultCodes.Core.SystemError, err.Error())
	}
	defer rows.Close()

	return parseRankingList(rows, int(rankingOrderParam.Length))
}

func initGetNearbyFriendsRankingsStmt() error {
	stmt, err := Database.Prepare(`
		WITH scores AS (
			WITH cat AS (
				SELECT golf_scoring FROM ranking.categories WHERE category = $1
			)
			SELECT
				*,
				CASE WHEN $2 THEN
					RANK() OVER (ORDER BY 
						CASE WHEN cat.golf_scoring THEN ranking.scores.score END DESC,
						CASE WHEN NOT cat.golf_scoring THEN ranking.scores.score END ASC
					)
				ELSE
					ROW_NUMBER() OVER (ORDER BY 
						CASE WHEN cat.golf_scoring THEN ranking.scores.score END DESC,
						CASE WHEN NOT cat.golf_scoring THEN ranking.scores.score END ASC
					)
				END AS ord
			FROM cat, ranking.scores
			WHERE
			    /* $7: requester pid; $8: friends list */
			    (owner_pid = ANY($8) OR owner_pid = $7) AND
				category = $1 AND
				/* $3: GroupIndex; $4: GroupNum */
				CASE WHEN $3 < 2 AND length(groups) >= 2 THEN get_byte(groups, $3) = $4 ELSE TRUE END
		), central_user AS (
			SELECT
				scores.ord,
				/* $6: LIMIT */
				GREATEST(scores.ord - $6 / 2, 1) AS min_ord,
				LEAST(scores.ord + $6 / 2, (SELECT max(scores.ord) FROM scores)) AS max_ord
			FROM scores
			WHERE owner_pid = $7
		)
		SELECT
			scores.unique_id,
			scores.owner_pid,
			scores.category,
			scores.groups,
			scores.score,
			scores.param,
			scores.update_date,
			COALESCE(ranking.common_data.data, ''::bytea),
			scores.ord,
			/* highly unfortunate */
			COUNT(*) OVER () AS count
		FROM central_user, scores
			LEFT OUTER JOIN ranking.common_data
			ON scores.unique_id = ranking.common_data.unique_id
			AND scores.owner_pid = ranking.common_data.owner_pid
		WHERE
			scores.ord >= central_user.min_ord AND
			scores.ord <= central_user.max_ord
		ORDER BY scores.ord
		OFFSET $5
		LIMIT $6
	`)
	if err != nil {
		return err
	}

	getNearbyFriendsRankings = stmt
	return nil
}
