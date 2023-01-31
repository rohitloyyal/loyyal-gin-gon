package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/gin-gonic/gin"

	"github.com/loyyal/loyyal-be-contract/controllers"
	"github.com/loyyal/loyyal-be-contract/nats"
	"github.com/loyyal/loyyal-be-contract/services"
)

var (
	version = "dev"
	service = "api"

	logger       *log.Logger
	server       *gin.Engine
	queueService *nats.Client

	contractController    controllers.ContractController
	authController        controllers.AuthController
	identityController    controllers.IdentityController
	walletController      controllers.WalletController
	transactionController controllers.TransactionController

	userService        services.UserService
	identityService    services.IdentityService
	walletService      services.WalletService
	transactionService services.TransactionService
	contractService    services.ContractService

	ctx     context.Context
	cluster *gocb.Cluster
	bucket  *gocb.Bucket
	err     error
)

func init() {
	ctx = context.TODO()
	logger = log.New(os.Stderr, fmt.Sprintf("api[%s]: ", version), log.Llongfile|log.Lmicroseconds|log.LstdFlags)
	//loading environments from file
	// err := godotenv.Load()
	// if err != nil {
	// 	logger.Fatal("error: while loading .env file")
	// }

	//initalising database connection
	logger.Println("couchbase connection pending")
	connectionString := os.Getenv("COUCHBASE_CONNECTION_URI")
	bucketName := os.Getenv("COUCHBASE_DEFAULT_BUCKET")

	username := os.Getenv("COUCHBASE_USERNAME")
	password := os.Getenv("COUCHBASE_PASSWORD")

	if connectionString == "" || bucketName == "" || username == "" || password == "" {
		logger.Fatal("error: missing environment configuration")
	}

	// For a secure cluster connection, use `couchbases://<your-cluster-ip>` instead.
	cluster, err = gocb.Connect("couchbase://"+connectionString, gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: username,
			Password: password,
		},
	})

	if err != nil {
		logger.Fatal(err.Error())
	}

	bucket = cluster.Bucket(bucketName)

	err = bucket.WaitUntilReady(5*time.Second, nil)
	if err != nil {
		logger.Fatalf("error initializing couchbase connection: %v", err)
	}

	logger.Println("couchbase connection established")

	// nats connection
	logger.Println("nats connection pending")

	natsUrl := os.Getenv("NATS_CONNECTION_URL")
	queueService, err := nats.NewClient(natsUrl)
	if err != nil {
		logger.Fatalf("error initializing NATS connection: %v", err)
	}
	logger.Println("nats connection established")

	// controller and service
	userService = services.NewUserService(bucket, ctx)
	identityService = services.NewIdentity(bucket, ctx)
	walletService = services.NewWallet(bucket, ctx)
	transactionService = services.NewTransaction(bucket, ctx)
	contractService = services.NewContract(bucket, ctx)

	authController = controllers.NewAuthController(userService)
	identityController = controllers.NewIdentityController(identityService)
	walletController = controllers.NewWallet(walletService, queueService)
	transactionController = controllers.NewTransactionController(transactionService, queueService)
	contractController = controllers.NewContractController(contractService)

	server = gin.Default()
}

func main() {
	logger.Println("starting...")
	defer cluster.Close(nil)

	basepath := server.Group("/v1")
	authController.AuthRoutes(basepath)
	identityController.IdentityRoutes(basepath)
	walletController.WalletRoutes(basepath)
	transactionController.TransactionRoutes(basepath)
	contractController.ContractRoutes(basepath)
	server.Run()
}
