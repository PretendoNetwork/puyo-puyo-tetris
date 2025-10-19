package nex

import (
	"fmt"
	"os"
	"strconv"

	common_globals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/puyo-puyo-tetris/globals"
)

func StartSecureServer() {
	globals.SecureServer = nex.NewPRUDPServer()
	globals.SecureServer.ByteStreamSettings.UseStructureHeader = true

	globals.SecureEndpoint = nex.NewPRUDPEndPoint(1)
	globals.SecureEndpoint.IsSecureEndPoint = true
	globals.SecureEndpoint.ServerAccount = globals.SecureServerAccount
	globals.SecureEndpoint.AccountDetailsByPID = globals.AccountDetailsByPID
	globals.SecureEndpoint.AccountDetailsByUsername = globals.AccountDetailsByUsername
	globals.SecureServer.BindPRUDPEndPoint(globals.SecureEndpoint)

	globals.SecureServer.LibraryVersions.SetDefault(nex.NewLibraryVersion(3, 5, 0))
	globals.SecureServer.AccessKey = "4eb0ca36"

	globals.SecureEndpoint.OnData(func(packet nex.PacketInterface) {
		request := packet.RMCMessage()

		fmt.Println("==Puyo Puyo Tetris - Secure==")
		fmt.Printf("Protocol ID: %d\n", request.ProtocolID)
		fmt.Printf("Method ID: %d\n", request.MethodID)
		fmt.Println("===============")
	})

	globals.SecureEndpoint.OnError(func(err *nex.Error) {
		globals.Logger.Errorf("Secure: %v", err)
	})

	globals.MatchmakingManager = common_globals.NewMatchmakingManager(globals.SecureEndpoint, globals.Postgres)
	globals.MatchmakingManager.GetUserFriendPIDs = globals.GetUserFriendPIDs
	globals.DatastoreManager = common_globals.NewDataStoreManager(globals.SecureEndpoint, globals.Postgres)
	globals.DatastoreManager.SetS3Config(globals.S3Bucket, globals.S3KeyBase, globals.S3Manager)
	globals.DatastoreManager.GetUserFriendPIDs = globals.GetUserFriendPIDs

	registerCommonSecureServerProtocols()

	port, _ := strconv.Atoi(os.Getenv("PN_PUYOPUYOTETRIS_SECURE_SERVER_PORT"))

	globals.SecureServer.Listen(port)
}
