package services

import (
	"context"
	"errors"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/utils/common"
)

type ContractService struct {
	bucket *gocb.Bucket
	ctx    context.Context
}

func NewContract(bucket *gocb.Bucket, ctx context.Context) ContractService {
	return ContractService{bucket: bucket, ctx: ctx}
}

func (service *ContractService) CreateContract(contract *models.Contract, creator string, channel string) error {
	col := service.bucket.DefaultCollection()
	contract.Identifier = common.GenerateIdentifier(30)
	contract.Creator = creator
	contract.Channel = channel
	contract.Status = "active"
	contract.CreatedAt = time.Now()
	contract.LastUpdatedAt = time.Now()
	contract.LastUpdatedBy = contract.Creator
	_, err := col.Insert(contract.Identifier, contract, nil)

	return err

}

func (service *ContractService) GetContract(contractId string) (any, error) {
	col := service.bucket.DefaultCollection()
	doc, err := col.Get(contractId, nil)
	if doc == nil {
		if err != nil {
			return "", errors.New("error: no contract found")
		}
	}
	return doc, err

}

func (service *WalletService) DeleteContract(walletId string, sessionedUser string) error {
	col := service.bucket.DefaultCollection()
	doc, err := col.Get(wallet_prefix+"/"+walletId, nil)
	if doc == nil {
		if err != nil {
			return errors.New("error: no contract found")
		}
	}

	var wallet models.Wallet
	err = doc.Content(&wallet)
	if err != nil {
		return err
	}

	wallet.IsDeleted = true
	wallet.LastUpdatedAt = time.Now()
	wallet.LastUpdatedBy = sessionedUser

	// TODO: need to convert it into soft delete
	_, err = col.Replace(wallet_prefix+"/"+walletId, wallet, nil)
	return err

}
