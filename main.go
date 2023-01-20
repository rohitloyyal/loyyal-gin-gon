package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	// "github.com/loyyal/golo/nats"

	"github.com/loyyal/loyyal-be-contract/controllers"
	"github.com/loyyal/loyyal-be-contract/services"
)

var (
	version = "dev"
	service = "api"

	logger *log.Logger
	server *gin.Engine

	contractController controllers.ContractController
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
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("error: while loading .env file")
	}

	//initalising database connection
	logger.Println("couchbase connection pending")
	connectionString := "localhost"
	bucketName := "testbucket"

	username := "Administrator"
	password := "password"

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
		logger.Fatal(err.Error())
	}

	logger.Println("couchbase connection established")

	// nats connection

	// controller and service
	contractService = services.New(bucket, ctx)
	contractController = controllers.New(contractService)

	server = gin.Default()
}

func main() {
	logger.Println("starting...")
	defer cluster.Close(nil)

	basepath := server.Group("/v1")
	contractController.RegisterRoutes(basepath)
	server.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
