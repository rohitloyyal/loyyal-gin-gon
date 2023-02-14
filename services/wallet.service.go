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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type WalletService struct {
	cluster *gocb.Cluster
	bucket  *gocb.Bucket
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

func NewWallet(cluster *gocb.Cluster, bucket *gocb.Bucket) WalletService {
	return WalletService{cluster: cluster, bucket: bucket}
}

func (service *WalletService) Create(ctx context.Context, wallet *models.Wallet, linkedTo string, preLoadAmount int64) error {
	wallet.Identifier = common.GenerateIdentifier(30)
	wallet.Channel = "loyyalchannel"
	wallet.DocType = "wallet"
	wallet.LinkedTo = []string{linkedTo}
	wallet.Balance = preLoadAmount

	wallet.WalletType = "regular"
	wallet.Creator = "admin"
	wallet.Status = ACTIVE

	location, _ := time.LoadLocation("UTC")
	now, _ := time.Parse(time.RFC1123, time.Now().In(location).Format(time.RFC1123))
	wallet.CreatedAt = now
	wallet.LastUpdatedAt = now
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

func (service *WalletService) Get(ctx context.Context, walletId string) (models.Wallet, error) {
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

func (service *WalletService) Update(ctx context.Context, walletId string, sessionedUser string, updatedStatus string) error {
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
	location, _ := time.LoadLocation("UTC")
	now, _ := time.Parse(time.RFC1123, time.Now().In(location).Format(time.RFC1123))
	wallet.LastUpdatedAt = now
	wallet.LastUpdatedBy = sessionedUser

	// TODO: need to convert it into soft delete
	_, err = col.Replace(wallet_prefix+"/"+walletId, wallet, nil)
	return err

}

func (service *WalletService) Delete(ctx context.Context, walletId string, sessionedUser string) error {
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
	location, _ := time.LoadLocation("UTC")
	now, _ := time.Parse(time.RFC1123, time.Now().In(location).Format(time.RFC1123))
	wallet.LastUpdatedAt = now
	wallet.LastUpdatedBy = sessionedUser

	// TODO: need to convert it into soft delete
	_, err = col.Replace(wallet_prefix+"/"+walletId, wallet, nil)
	return err

}

func (service *WalletService) Filter(ctx context.Context, queryString string, params map[string]interface{}, sortBy string, limit int) ([]*models.Wallet, error) {
	fName := "service/wallet/filter"
	tracer := otel.Tracer("walletFilter")
	_, span := tracer.Start(ctx, fName)
	defer span.End()

	// TODO: we can even make the retuning fields as input from calling methods insead of returning all fields
	query := "select data.* from `testbucket` data where type='wallet' "
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

	return parseWalletRows(rows), nil
}

/*
Implemented for the linked wallets only
*/
func (service *WalletService) CustomFilterQuery(ctx context.Context, selector string, queryString string, params map[string]interface{}, sortBy string, limit int) ([]*models.Wallet, error) {
	fName := "service/transaction/customFilterQuery"
	tracer := otel.Tracer("api")
	_, span := tracer.Start(ctx, fName)
	defer span.End()

	// TODO: we can even make the retuning fields as input from calling methods insead of returning all fields
	query := "select " + selector + " from `testbucket` where type='wallet' "
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
