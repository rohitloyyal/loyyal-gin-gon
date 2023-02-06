package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/nats"
	"github.com/loyyal/loyyal-be-contract/utils/common"
)

type WalletService struct {
	cluster *gocb.Cluster
	bucket  *gocb.Bucket
	ctx     context.Context
}

const (
	wallet_prefix = "wallet"
)

const (
	ACTIVE     = "active"
	SUBSPENDED = "suspended"
	DISABLED   = "disabled"
	EXPIRED    = "expired"
)

func NewWallet(cluster *gocb.Cluster, bucket *gocb.Bucket, ctx context.Context) WalletService {
	return WalletService{cluster: cluster, bucket: bucket, ctx: ctx}
}

func (service *WalletService) Create(wallet *models.Wallet, linkedTo string, preLoadAmount int64) error {
	wallet.Identifier = common.GenerateIdentifier(30)
	wallet.Channel = "loyyalchannel"
	wallet.DocType = "wallet"
	wallet.LinkedTo = []string{linkedTo}
	wallet.Balance = preLoadAmount

	wallet.WalletType = "regular"
	wallet.Creator = "admin"
	wallet.Status = ACTIVE

	wallet.CreatedAt = time.Now()
	wallet.LastUpdatedAt = time.Now()
	wallet.LastUpdatedBy = wallet.Creator

	// adding this reference that will be send to blockchain
	// and used in callback to update the same document in database
	ref, err := common.NewRefID()
	if err != nil {
		return err
	}
	wallet.Ref = ref

	col := service.bucket.DefaultCollection()
	_, err = col.Insert(wallet_prefix+"/"+wallet.Identifier, wallet, nil)
	return err
}

func publishWalletCreate(ctx context.Context, nats *nats.Client, wallet *models.Wallet) {
	if err := nats.Publish(ctx, &models.CreateRequest{RefID: wallet.Ref, Amount: wallet.Balance, Channel: wallet.Channel}); err != nil {
		fmt.Print("failed to write wallet to NATS (failing over to retry service): %w", err)
	}
}

func (service *WalletService) Get(walletId string) (models.Wallet, error) {
	col := service.bucket.DefaultCollection()
	doc, err := col.Get(wallet_prefix+"/"+walletId, nil)
	if doc == nil {
		if err != nil {
			return models.Wallet{}, errors.New("error: no wallet found")
		}
	}

	var wallet models.Wallet
	doc.Content(&wallet)
	if err != nil {
		return models.Wallet{}, err
	}

	return wallet, err

}

func (service *WalletService) Update(walletId string, sessionedUser string, updatedStatus string) error {
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

	wallet.Status = updatedStatus
	wallet.LastUpdatedAt = time.Now()
	wallet.LastUpdatedBy = sessionedUser

	// TODO: need to convert it into soft delete
	_, err = col.Replace(wallet_prefix+"/"+walletId, wallet, nil)
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

func (service *WalletService) Filter(queryString string, params map[string]interface{}, sortBy string, limit int) ([]*models.Wallet, error) {

	// TODO: we can even make the retuning fields as input from calling methods insead of returning all fields
	query := "select data.* from `testbucket` data where type='wallet' "
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

	return parseWalletRows(rows), nil
}

/*
Implemented for the linked wallets only
*/
func (service *WalletService) CustomFilterQuery(selector string, queryString string, params map[string]interface{}, sortBy string, limit int) ([]*models.Wallet, error) {

	// TODO: we can even make the retuning fields as input from calling methods insead of returning all fields
	query := "select " + selector + " from `testbucket` where type='wallet' "
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

	return parseWalletRows(rows), nil
}

func parseWalletRows(rows *gocb.QueryResult) []*models.Wallet {
	var wallets []*models.Wallet
	for rows.Next() {
		var obj models.Wallet
		err := rows.Row(&obj)
		if err != nil {
			panic(err)
		}
		wallets = append(wallets, &obj)
	}
	defer rows.Close()
	err := rows.Err()
	if err != nil {
		panic(err)
	}

	return wallets
}
