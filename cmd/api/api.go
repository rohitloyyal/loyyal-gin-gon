package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/gin-gonic/gin"

	"github.com/loyyal/loyyal-be-contract/controllers"
	"github.com/loyyal/loyyal-be-contract/middleware"
	"github.com/loyyal/loyyal-be-contract/nats"
	"github.com/loyyal/loyyal-be-contract/services"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var (
	version = "dev"
	service = "api"

	logger       *log.Logger
	server       *gin.Engine
	queueService *nats.Client

	commonController      controllers.CommonController
	contractController    controllers.ContractController
	authController        controllers.AuthController
	identityController    controllers.IdentityController
	walletController      controllers.WalletController
	transactionController controllers.TransactionController

	authService        services.AuthService
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

	connectionString := os.Getenv("COUCHBASE_CONNECTION_URI")
	bucketName := os.Getenv("COUCHBASE_DEFAULT_BUCKET")

	database_username := os.Getenv("COUCHBASE_USERNAME")
	database_password := os.Getenv("COUCHBASE_PASSWORD")

	bootstrap_username := os.Getenv("BOOTSTRAP_USERNAME")
	bootstrap_password := os.Getenv("BOOTSTRAP_PASSWORD")

	if connectionString == "" || bucketName == "" || database_username == "" ||
		database_password == "" || bootstrap_username == "" || bootstrap_password == "" {
		logger.Fatal("error: missing environment configuration")
	}

	//initalising database connection
	logger.Println("couchbase connection pending")
	// For a secure cluster connection, use `couchbases://<your-cluster-ip>` instead.
	cluster, err = gocb.Connect("couchbase://"+connectionString, gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: database_username,
			Password: database_password,
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
	authService = services.NewAuthService(cluster, bucket)
	identityService = services.NewIdentity(cluster, bucket)
	walletService = services.NewWallet(cluster, bucket)
	transactionService = services.NewTransaction(cluster, bucket)
	contractService = services.NewContract(cluster, bucket)

	authController = controllers.NewAuthController(identityService)
	identityController = controllers.NewIdentityController(logger, identityService, walletService, queueService)
	walletController = controllers.NewWallet(walletService, transactionService, queueService)
	transactionController = controllers.NewTransactionController(logger, transactionService, contractService, walletService, queueService)
	contractController = controllers.NewContractController(contractService)

	// create bootstrap identity
	err = identityService.CreateBootstrapIdentity(ctx, bootstrap_username, bootstrap_password)
	if err != nil {
		logger.Fatalf("bootstrap errors: %v", err)
	}
	logger.Println("admin identity bootstraped")

	// tracing
	tp, err := JaegerTraceProvider()
	if err != nil {
		logger.Fatalf("tracing errors: %v", err)
	}

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// initialising server
	server = gin.Default()
}

func JaegerTraceProvider() (*sdktrace.TracerProvider, error) {
	jaeger_agent_addr := os.Getenv("JAEGER_COLLECTION_URL")
	if jaeger_agent_addr == "" {
		return nil, errors.New("no tracing collection endpoint given")
	}
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaeger_agent_addr + "/api/traces")))
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
			semconv.DeploymentEnvironmentKey.String(version),
		)),
	)
	return tp, nil

}
func main() {
	logger.Println("starting...")
	defer cluster.Close(nil)

	server.Use(middleware.CORSMiddleware())
	server.Use(otelgin.Middleware("api"))

	basepath := server.Group("/v1")
	commonController.CommonRoutes(basepath)
	authController.AuthRoutes(basepath)
	identityController.IdentityRoutes(basepath)
	walletController.WalletRoutes(basepath)
	transactionController.TransactionRoutes(basepath)
	contractController.ContractRoutes(basepath)
	server.Run()
}
