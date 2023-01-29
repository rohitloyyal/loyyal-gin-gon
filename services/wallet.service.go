package services

import (
	"context"
	"errors"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/utils/common"
)

type WalletService struct {
	bucket *gocb.Bucket
	ctx    context.Context
}

const (
	wallet_prefix = "wallet"
)

func NewWallet(bucket *gocb.Bucket, ctx context.Context) WalletService {
	return WalletService{bucket: bucket, ctx: ctx}
}

func (service *WalletService) Create(wallet *models.Wallet, linkedTo string, preLoadAmount int64) error {
	wallet.Identifier = common.GenerateIdentifier(30)
	wallet.Channel = "loyyalchannel"
	wallet.DocType = "wallet"
	wallet.LinkedTo = []string{linkedTo}
	wallet.Balance = preLoadAmount

	wallet.WalletType = "regular"
	wallet.Creator = "consumer"
	wallet.Status = "active"

	wallet.CreatedAt = time.Now()
	wallet.LastUpdatedAt = time.Now()
	wallet.LastUpdatedBy = wallet.Creator

	col := service.bucket.DefaultCollection()
	_, err := col.Insert(wallet_prefix+"/"+wallet.Identifier, wallet, nil)

	return err

}

func (service *WalletService) Get(walletId string) (any, error) {
	col := service.bucket.DefaultCollection()
	doc, err := col.Get(wallet_prefix+"/"+walletId, nil)
	if doc == nil {
		if err != nil {
			return "", errors.New("error: no wallet found")
		}
	}

	var wallet models.Wallet
	doc.Content(&wallet)
	if err != nil {
		return nil, err
	}

	return doc, err

}

func (service *WalletService) Filter(contract *models.Wallet) error {
	col := service.bucket.DefaultCollection()
	_, err := col.Upsert(contract.Identifier, contract, nil)

	return err

}

func (service *WalletService) Delete(walletId string, sessionedUser string) error {
	col := service.bucket.DefaultCollection()
	doc, err := col.Get(wallet_prefix+"/"+walletId, nil)
	if doc == nil {
		if err != nil {
			return errors.New("error: no wallet found")	
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
