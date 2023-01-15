package intializers

import (
	"log"
	"time"

	"github.com/couchbase/gocb/v2"
)

var DB *gocb.Bucket
var Cluster *gocb.Cluster

func ConnectToDB() {
	log.Println("couchbase connection pending")
	var err error

	// Uncomment following line to enable logging
	// gocb.SetLogger(gocb.VerboseStdioLogger())

	// Update this to your cluster details
	connectionString := "localhost"
	bucketName := "testbucket"

	username := "admin"
	password := "3M9Oh4Hq1qE"

	// tiemout := "60"

	// For a secure cluster connection, use `couchbases://<your-cluster-ip>` instead.
	cluster, err := gocb.Connect("couchbase://"+connectionString, gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: username,
			Password: password,
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	bucket := cluster.Bucket(bucketName)

	err = bucket.WaitUntilReady(5*time.Second, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("couchbase connection established")
	// Get a reference to the default collection, required for older Couchbase server versions
	// col := bucket.DefaultCollection()
	Cluster = cluster
	DB = bucket
}

// func Close() error {
// 	return DB
// }
