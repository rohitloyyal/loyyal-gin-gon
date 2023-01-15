package api

import "github.com/couchbase/gocb/v2"

type Handler struct {
	DB *gocb.Bucket
	// NATS *nats.Client
}
