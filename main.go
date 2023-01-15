package main

import (
	"fmt"
	"log"
	"os"

	"github.com/couchbase/gocb/v2"
	"github.com/gin-gonic/gin"

	// "github.com/loyyal/golo/nats"
	"github.com/loyyal/loyyal-be-contract/api"
	"github.com/loyyal/loyyal-be-contract/controllers"
	"github.com/loyyal/loyyal-be-contract/intializers"
)

type Handler struct {
	version string
	DB      *gocb.Bucket
	// NATS *nats.Client
}

var (
	// Set this with compile arg, e.g.:
	// go build -ldflags "-X main.version=1.0.5-5+$(git rev-parse --short HEAD)" .
	version = "dev"

	service = "api"

	// Metrics
	// mHits = behold.NewCounter("golo-cmd-api-hits", "The number of hits recieved", "1")
)

func init() {
	intializers.LoadEnv()
	intializers.ConnectToDB()
	// intializers.ConnectNats()
}

func main() {
	logger := log.New(os.Stderr, fmt.Sprintf("api[%s]: ", version), log.Llongfile|log.Lmicroseconds|log.LstdFlags)
	logger.Println("starting...")

	r := gin.Default()


	// h := &Handler{
	// 	DB: bucket,
	// }

	// create NATS connection

	r.GET("/ping", api.Ping)
	r.GET("/connect")

	r.POST("/contract", controllers.ContractCreate)
	r.GET("/contract", controllers.ContractGet)
	r.GET("/contracts", api.Ping)
	r.DELETE("/contract", api.Ping)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
