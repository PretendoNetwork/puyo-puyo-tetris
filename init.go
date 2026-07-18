package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	commonglobals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"

	pbaccount "github.com/PretendoNetwork/grpc/go/account"
	pbfriends "github.com/PretendoNetwork/grpc/go/friends"
	"github.com/PretendoNetwork/nex-go/v2"
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
	s3Bucket := os.Getenv("PN_PUYOPUYOTETRIS_CONFIG_S3_BUCKET")
	s3KeyBase := os.Getenv("PN_PUYOPUYOTETRIS_CONFIG_S3_KEY_BASE")
	s3Insecure := os.Getenv("PN_PUYOPUYOTETRIS_CONFIG_S3_INSECURE")

	kerberosPassword := os.Getenv("PN_PUYOPUYOTETRIS_KERBEROS_PASSWORD")
	authenticationServerPort := os.Getenv("PN_PUYOPUYOTETRIS_AUTHENTICATION_SERVER_PORT")
	secureServerHost := os.Getenv("PN_PUYOPUYOTETRIS_SECURE_SERVER_HOST")
	secureServerPort := os.Getenv("PN_PUYOPUYOTETRIS_SECURE_SERVER_PORT")
	accountGRPCHost := os.Getenv("PN_PUYOPUYOTETRIS_ACCOUNT_GRPC_HOST")
	accountGRPCPort := os.Getenv("PN_PUYOPUYOTETRIS_ACCOUNT_GRPC_PORT")
	accountGRPCAPIKey := os.Getenv("PN_PUYOPUYOTETRIS_ACCOUNT_GRPC_API_KEY")
	friendsGRPCHost := os.Getenv("PN_PUYOPUYOTETRIS_FRIENDS_GRPC_HOST")
	friendsGRPCPort := os.Getenv("PN_PUYOPUYOTETRIS_FRIENDS_GRPC_PORT")
	friendsGRPCAPIKey := os.Getenv("PN_PUYOPUYOTETRIS_FRIENDS_GRPC_API_KEY")
	localAuthMode := os.Getenv("PN_PUYOPUYOTETRIS_LOCAL_AUTH")
	healthCheckPort := os.Getenv("PN_PUYOPUYOTETRIS_HEALTH_CHECK_PORT")

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

	globals.GRPCAccountClient = pbaccount.NewAccountClient(globals.GRPCAccountClientConnection)
	globals.GRPCAccountCommonMetadata = metadata.Pairs(
		"X-API-Key", accountGRPCAPIKey,
	)

	globals.GRPCFriendsClientConnection, err = grpc.Dial(fmt.Sprintf("%s:%s", friendsGRPCHost, friendsGRPCPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		globals.Logger.Criticalf("Failed to connect to account gRPC server: %v", err)
		os.Exit(0)
	}

	globals.GRPCFriendsClient = pbfriends.NewFriendsClient(globals.GRPCFriendsClientConnection)
	globals.GRPCFriendsCommonMetadata = metadata.Pairs(
		"X-API-Key", friendsGRPCAPIKey,
	)

	globals.LocalAuthMode = localAuthMode == "1"
	if globals.LocalAuthMode {
		globals.Logger.Warning("Local authentication mode is enabled. Token validation will be skipped!")
		globals.Logger.Warning("This is insecure and could allow ban bypasses!")
	}

	secure := s3Insecure != "1"
	if !secure {
		globals.Logger.Warning("S3 is set to use HTTP! This is insecure.")
	}

	staticCredentials := credentials.NewStaticV4(s3AccessKey, s3AccessSecret, "")

	minIOClient, err := minio.New(s3Endpoint, &minio.Options{
		Creds:  staticCredentials,
		Secure: secure,
	})

	if err != nil {
		panic(err)
	}

	globals.MinIOClient = minIOClient
	globals.S3Manager = commonglobals.NewMinIOManager(minIOClient)
	globals.S3Bucket = s3Bucket
	globals.S3KeyBase = s3KeyBase

	globals.Postgres, err = sql.Open("postgres", os.Getenv("PN_PUYOPUYOTETRIS_POSTGRES_URI"))
	if err != nil {
		globals.Logger.Critical(err.Error())
	}
	globals.Logger.Success("Connected to Postgres!")

	if strings.TrimSpace(healthCheckPort) == "" {
		globals.Logger.Warning("Basic UDP health check will not be enabled. PN_PUYOPUYOTETRIS_HEALTH_CHECK_PORT environment variable not set")
	} else if port, err := strconv.Atoi(healthCheckPort); err != nil {
		globals.Logger.Errorf("PN_PUYOPUYOTETRIS_HEALTH_CHECK_PORT is not a valid port. Expected 0-65535, got %s", healthCheckPort)
		os.Exit(0)
	} else if port < 0 || port > 65535 {
		globals.Logger.Errorf("PN_PUYOPUYOTETRIS_HEALTH_CHECK_PORT is not a valid port. Expected 0-65535, got %s", healthCheckPort)
		os.Exit(0)
	} else {
		nex.EnableBasicUDPHealthCheck(port)
	}
}
