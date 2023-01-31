package services

import (
	"context"
	"errors"

	"github.com/couchbase/gocb/v2"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/utils/common"
)

type TransactionService struct {
	bucket *gocb.Bucket
	ctx    context.Context
}

const (
	transaction_prefix = "tx"
)

func NewTransaction(bucket *gocb.Bucket, ctx context.Context) TransactionService {
	return TransactionService{bucket: bucket, ctx: ctx}
}

func (service *TransactionService) Create(transaction *models.Transaction) error {
	transaction.DocType = "tx"
	transaction.Channel = "loyyalchannel"
	transaction.Creator = "admin"
	transaction.TransactionExtID.Ref = common.GenerateIdentifier(62)
	transaction.Ref = common.GenerateIdentifier(30)

	col := service.bucket.DefaultCollection()
	_, err := col.Insert(transaction_prefix+"/"+transaction.TransactionExtID.Ref, transaction, nil)

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

func (service *TransactionService) Filter(contract *models.Wallet) error {
	col := service.bucket.DefaultCollection()
	_, err := col.Upsert(contract.Identifier, contract, nil)

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
