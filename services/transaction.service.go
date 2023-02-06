package services

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/utils/common"
)

type TransactionService struct {
	cluster *gocb.Cluster
	bucket  *gocb.Bucket
	ctx     context.Context
}

const (
	transaction_prefix = "tx"
)

func NewTransaction(cluster *gocb.Cluster, bucket *gocb.Bucket, ctx context.Context) TransactionService {
	return TransactionService{cluster: cluster, bucket: bucket, ctx: ctx}
}

func (service *TransactionService) Create(transaction *models.Transaction) error {
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
	return err

}

func (service *TransactionService) Delete(txId string, sessionedUser string) error {
	col := service.bucket.DefaultCollection()
	doc, err := col.Get(transaction_prefix+"/"+txId, nil)
	if doc == nil {
		if err != nil {
			return errors.New("error: no wallet found")
		}
	}

	var transaction models.Transaction
	err = doc.Content(&transaction)
	if err != nil {
		return err
	}

	// transaction.IsDeleted = true
	// transaction.LastUpdatedAt = time.Now()
	// transaction.LastUpdatedBy = sessionedUser

	// TODO: need to convert it into soft delete
	_, err = col.Replace(transaction_prefix+"/"+txId, transaction, nil)
	return err

}

func (service *TransactionService) Get(transactionId string) (any, error) {
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

func (service *TransactionService) Filter(queryString string, params map[string]interface{}, sortBy string, limit int) ([]*models.Transaction, error) {

	// TODO: we can even make the retuning fields as input from calling methods insead of returning all fields
	query := "select data.* from `testbucket`.`_default`.`_default` data where type='tx' "
	query += queryString
	query += " order by " + sortBy
	if limit != -1 {
		query += " limit " + strconv.Itoa(limit)
	}

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
