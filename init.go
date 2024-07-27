package main

import (
	"fmt"
	"github.com/PretendoNetwork/nex-protocols-common-go/v2/datastore"
	puyodatastore "github.com/PretendoNetwork/puyo-puyo-tetris/datastore"
	"os"
	"strconv"
	"strings"

	pb "github.com/PretendoNetwork/grpc-go/account"
	"github.com/PretendoNetwork/plogger-go"
	"github.com/PretendoNetwork/puyo-puyo-tetris/globals"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func init() {
	globals.Logger = plogger.NewLogger()

	var err error

	err = godotenv.Load()
	if err != nil {
		globals.Logger.Warning("Error loading .env file")
	}

	s3Endpoint := os.Getenv("PN_PUYOPUYOTETRIS_CONFIG_S3_ENDPOINT")
	s3AccessKey := os.Getenv("PN_PUYOPUYOTETRIS_CONFIG_S3_ACCESS_KEY")
	s3AccessSecret := os.Getenv("PN_PUYOPUYOTETRIS_CONFIG_S3_ACCESS_SECRET")

	kerberosPassword := os.Getenv("PN_PUYOPUYOTETRIS_KERBEROS_PASSWORD")
	authenticationServerPort := os.Getenv("PN_PUYOPUYOTETRIS_AUTHENTICATION_SERVER_PORT")
	secureServerHost := os.Getenv("PN_PUYOPUYOTETRIS_SECURE_SERVER_HOST")
	secureServerPort := os.Getenv("PN_PUYOPUYOTETRIS_SECURE_SERVER_PORT")
	accountGRPCHost := os.Getenv("PN_PUYOPUYOTETRIS_ACCOUNT_GRPC_HOST")
	accountGRPCPort := os.Getenv("PN_PUYOPUYOTETRIS_ACCOUNT_GRPC_PORT")
	accountGRPCAPIKey := os.Getenv("PN_PUYOPUYOTETRIS_ACCOUNT_GRPC_API_KEY")

	if strings.TrimSpace(kerberosPassword) == "" {
		globals.Logger.Warningf("PN_PUYOPUYOTETRIS_KERBEROS_PASSWORD environment variable not set. Using default password: %q", globals.KerberosPassword)
	} else {
		globals.KerberosPassword = kerberosPassword
	}

	globals.InitAccounts()

	if strings.TrimSpace(authenticationServerPort) == "" {
		globals.Logger.Error("PN_PUYOPUYOTETRIS_AUTHENTICATION_SERVER_PORT environment variable not set")
		os.Exit(0)
	}

	if port, err := strconv.Atoi(authenticationServerPort); err != nil {
		globals.Logger.Errorf("PN_PUYOPUYOTETRIS_AUTHENTICATION_SERVER_PORT is not a valid port. Expected 0-65535, got %s", authenticationServerPort)
		os.Exit(0)
	} else if port < 0 || port > 65535 {
		globals.Logger.Errorf("PN_PUYOPUYOTETRIS_AUTHENTICATION_SERVER_PORT is not a valid port. Expected 0-65535, got %s", authenticationServerPort)
		os.Exit(0)
	}

	if strings.TrimSpace(secureServerHost) == "" {
		globals.Logger.Error("PN_PUYOPUYOTETRIS_SECURE_SERVER_HOST environment variable not set")
		os.Exit(0)
	}

	if strings.TrimSpace(secureServerPort) == "" {
		globals.Logger.Error("PN_PUYOPUYOTETRIS_SECURE_SERVER_PORT environment variable not set")
		os.Exit(0)
	}

	if port, err := strconv.Atoi(secureServerPort); err != nil {
		globals.Logger.Errorf("PN_PUYOPUYOTETRIS_SECURE_SERVER_PORT is not a valid port. Expected 0-65535, got %s", secureServerPort)
		os.Exit(0)
	} else if port < 0 || port > 65535 {
		globals.Logger.Errorf("PN_PUYOPUYOTETRIS_SECURE_SERVER_PORT is not a valid port. Expected 0-65535, got %s", secureServerPort)
		os.Exit(0)
	}

	if strings.TrimSpace(accountGRPCHost) == "" {
		globals.Logger.Error("PN_PUYOPUYOTETRIS_ACCOUNT_GRPC_HOST environment variable not set")
		os.Exit(0)
	}

	if strings.TrimSpace(accountGRPCPort) == "" {
		globals.Logger.Error("PN_PUYOPUYOTETRIS_ACCOUNT_GRPC_PORT environment variable not set")
		os.Exit(0)
	}

	if port, err := strconv.Atoi(accountGRPCPort); err != nil {
		globals.Logger.Errorf("PN_PUYOPUYOTETRIS_ACCOUNT_GRPC_PORT is not a valid port. Expected 0-65535, got %s", accountGRPCPort)
		os.Exit(0)
	} else if port < 0 || port > 65535 {
		globals.Logger.Errorf("PN_PUYOPUYOTETRIS_ACCOUNT_GRPC_PORT is not a valid port. Expected 0-65535, got %s", accountGRPCPort)
		os.Exit(0)
	}

	if strings.TrimSpace(accountGRPCAPIKey) == "" {
		globals.Logger.Warning("Insecure gRPC server detected. PN_PUYOPUYOTETRIS_ACCOUNT_GRPC_API_KEY environment variable not set")
	}

	globals.GRPCAccountClientConnection, err = grpc.Dial(fmt.Sprintf("%s:%s", accountGRPCHost, accountGRPCPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		globals.Logger.Criticalf("Failed to connect to account gRPC server: %v", err)
		os.Exit(0)
	}

	globals.GRPCAccountClient = pb.NewAccountClient(globals.GRPCAccountClientConnection)
	globals.GRPCAccountCommonMetadata = metadata.Pairs(
		"X-API-Key", accountGRPCAPIKey,
	)

	staticCredentials := credentials.NewStaticV4(s3AccessKey, s3AccessSecret, "")

	minIOClient, err := minio.New(s3Endpoint, &minio.Options{
		Creds:  staticCredentials,
		Secure: true,
	})

	if err != nil {
		panic(err)
	}

	globals.MinIOClient = minIOClient
	globals.Presigner = datastore.NewS3Presigner(globals.MinIOClient)

	puyodatastore.ConnectPostgres()
}
