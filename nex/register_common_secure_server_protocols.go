package nex

import (
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	commondatastore "github.com/PretendoNetwork/nex-protocols-common-go/v2/datastore"
	commonranking "github.com/PretendoNetwork/nex-protocols-common-go/v2/ranking"
	commonsecure "github.com/PretendoNetwork/nex-protocols-common-go/v2/secure-connection"
	datastore "github.com/PretendoNetwork/nex-protocols-go/v2/datastore"
	ranking "github.com/PretendoNetwork/nex-protocols-go/v2/ranking"
	secure "github.com/PretendoNetwork/nex-protocols-go/v2/secure-connection"
	puyodatastore "github.com/PretendoNetwork/puyo-puyo-tetris/datastore"
	"github.com/PretendoNetwork/puyo-puyo-tetris/globals"
	puyoranking "github.com/PretendoNetwork/puyo-puyo-tetris/ranking"
	"os"

	commonmatchmaking "github.com/PretendoNetwork/nex-protocols-common-go/v2/match-making"
	commonmatchmakingext "github.com/PretendoNetwork/nex-protocols-common-go/v2/match-making-ext"
	commonmatchmakeextension "github.com/PretendoNetwork/nex-protocols-common-go/v2/matchmake-extension"
	matchmaking "github.com/PretendoNetwork/nex-protocols-go/v2/match-making"
	matchmakingext "github.com/PretendoNetwork/nex-protocols-go/v2/match-making-ext"
	matchmakeextension "github.com/PretendoNetwork/nex-protocols-go/v2/matchmake-extension"

	commonnattraversal "github.com/PretendoNetwork/nex-protocols-common-go/v2/nat-traversal"
	nattraversal "github.com/PretendoNetwork/nex-protocols-go/v2/nat-traversal"

	matchmakingtypes "github.com/PretendoNetwork/nex-protocols-go/v2/match-making/types"

	common_globals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
)

func CreateReportDBRecord(_ *types.PID, _ *types.PrimitiveU32, _ *types.QBuffer) error {
	return nil
}

// TO DO:
// How do clubs work?
// GetObjectInfoByDataID
// UpdateObjectPeriodByDataIDWithPassword
// UpdateObjectMetaBinaryByDataIDWithPassword
// UpdateObjectDataTypeByDataIDWithPassword

func registerCommonSecureServerProtocols() {
	secureProtocol := secure.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(secureProtocol)
	commonSecureProtocol := commonsecure.NewCommonProtocol(secureProtocol)

	commonSecureProtocol.CreateReportDBRecord = CreateReportDBRecord

	// Datastore - replays, user profiles (country, preferred character)
	datastoreProtocol := datastore.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(datastoreProtocol)
	commonDatastoreProtocol := commondatastore.NewCommonProtocol(datastoreProtocol)

	commonDatastoreProtocol.S3Bucket = os.Getenv("PN_PUYOPUYOTETRIS_CONFIG_S3_BUCKET")
	commonDatastoreProtocol.SetMinIOClient(globals.MinIOClient)
	commonDatastoreProtocol.GetObjectInfosByDataStoreSearchParam = puyodatastore.GetObjectInfosByDataStoreSearchParam
	commonDatastoreProtocol.InitializeObjectByPreparePostParam = puyodatastore.InitializeObjectByPreparePostParam
	commonDatastoreProtocol.InitializeObjectRatingWithSlot = puyodatastore.InitializeObjectRatingWithSlot
	commonDatastoreProtocol.GetObjectInfoByDataID = puyodatastore.GetObjectInfoByDataID
	commonDatastoreProtocol.UpdateObjectPeriodByDataIDWithPassword = puyodatastore.UpdateObjectPeriodByDataIDWithPassword
	commonDatastoreProtocol.UpdateObjectMetaBinaryByDataIDWithPassword = puyodatastore.UpdateObjectMetaBinaryByDataIDWithPassword
	commonDatastoreProtocol.UpdateObjectDataTypeByDataIDWithPassword = puyodatastore.UpdateObjectDataTypeByDataIDWithPassword
	commonDatastoreProtocol.GetObjectInfoByDataIDWithPassword = puyodatastore.GetObjectInfoByDataIDWithPassword
	commonDatastoreProtocol.GetObjectInfoByPersistenceTargetWithPassword = puyodatastore.GetObjectInfoByPersistenceTargetWithPassword

	// Ranking - ??
	rankingProtocol := ranking.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(rankingProtocol)
	commonRankingProtocol := commonranking.NewCommonProtocol(rankingProtocol)
	commonRankingProtocol.GetRankingsAndCountByCategoryAndRankingOrderParam = puyoranking.GetRankingsAndCountByCategoryAndRankingOrderParam

	// Matchmaking stuff - National Puzzle League
	natTraversalProtocol := nattraversal.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(natTraversalProtocol)
	commonnattraversal.NewCommonProtocol(natTraversalProtocol)

	matchMakingProtocol := matchmaking.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(matchMakingProtocol)
	commonmatchmaking.NewCommonProtocol(matchMakingProtocol)

	matchMakingExtProtocol := matchmakingext.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(matchMakingExtProtocol)
	commonmatchmakingext.NewCommonProtocol(matchMakingExtProtocol)

	matchmakeExtensionProtocol := matchmakeextension.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(matchmakeExtensionProtocol)
	commonMatchmakeExtensionProtocol := commonmatchmakeextension.NewCommonProtocol(matchmakeExtensionProtocol)

	commonMatchmakeExtensionProtocol.OnAfterAutoMatchmakeWithSearchCriteriaPostpone = func(packet nex.PacketInterface, lstSearchCriteria *types.List[*matchmakingtypes.MatchmakeSessionSearchCriteria], anyGathering *types.AnyDataHolder, strMessage *types.String) {
		for _, session := range common_globals.Sessions {
			globals.Logger.Info(session.GameMatchmakeSession.FormatToString(1))
		}
	}

}
