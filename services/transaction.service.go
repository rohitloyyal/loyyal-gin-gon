package services

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/utils/common"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type TransactionService struct {
	cluster *gocb.Cluster
	bucket  *gocb.Bucket
}

const (
	transaction_prefix = "tx"
)

func NewTransaction(cluster *gocb.Cluster, bucket *gocb.Bucket) TransactionService {
	return TransactionService{cluster: cluster, bucket: bucket}
}

func (service *TransactionService) Create(ctx context.Context, transaction *models.Transaction) error {
	fName := "service/transaction/create"
	tracer := otel.Tracer("api")
	_, span := tracer.Start(ctx, fName)
	defer span.End()

	span.AddEvent("creating transaction")

	transaction.DocType = "tx"
	transaction.Channel = "loyyalchannel"
	transaction.Creator = "admin"
	transaction.ExtID = common.GenerateIdentifier(62)
	transaction.CreatedOn = time.Now()

	// adding this reference that will be send to blockchain
	// and used in callback to update the same document in database
	ref, err := common.NewRefID()
	if err != nil {
		return err
	}
	transaction.RefID = ref

	col := service.bucket.DefaultCollection()
	_, err = col.Insert(transaction_prefix+"/"+transaction.ExtID, transaction, nil)
	span.AddEvent("added transaction to couchbase")

	return err

}

func (service *TransactionService) Get(ctx context.Context, transactionId string) (any, error) {
	fName := "service/transaction/get"
	tracer := otel.Tracer("api")
	_, span := tracer.Start(ctx, fName)
	defer span.End()

	col := service.bucket.DefaultCollection()
	doc, err := col.Get(transaction_prefix+"/"+transactionId, nil)
	if doc == nil {
		if err != nil {
			return "", errors.New("error: no transaction found")
		}
	}

	var transaction models.Transaction
	err = doc.Content(&transaction)
	if err != nil {
		return nil, err
	}
	return transaction, err

}

func (service *TransactionService) Filter(ctx context.Context, queryString string, params map[string]interface{}, sortBy string, limit int) ([]*models.Transaction, error) {
	fName := "service/transaction/create"
	tracer := otel.Tracer("api")
	_, span := tracer.Start(ctx, fName)
	defer span.End()

	// TODO: we can even make the retuning fields as input from calling methods insead of returning all fields
	query := "select data.* from `testbucket`.`_default`.`_default` data where type='tx' "
	query += queryString
	query += " order by " + sortBy
	if limit != -1 {
		query += " limit " + strconv.Itoa(limit)
	}

	span.SetAttributes(attribute.String("query", query))
	rows, err := service.cluster.Query(
		query,
		&gocb.QueryOptions{NamedParameters: params})

	if err != nil {
		return nil, err
	}

	return parseTransactionRows(rows), nil
}

func parseTransactionRows(rows *gocb.QueryResult) []*models.Transaction {
	var transactions []*models.Transaction
	for rows.Next() {
		var obj models.Transaction
		err := rows.Row(&obj)
		if err != nil {
			panic(err)
		}
		transactions = append(transactions, &obj)
	}
	defer rows.Close()
	err := rows.Err()
	if err != nil {
		panic(err)
	}

	return transactions
}
