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

type ContractService struct {
	cluster *gocb.Cluster
	bucket  *gocb.Bucket
}

const (
	contract_prefix = "contract"
)

const (
	CONTRACT_STATUS_ACTIVE  = "active"
	CONTRACT_STATUS_PENDING = "pending"
	CONTRACT_STATUS_EXPIRED = "expired"
)

func NewContract(cluster *gocb.Cluster, bucket *gocb.Bucket) ContractService {
	return ContractService{cluster: cluster, bucket: bucket}
}

func (service *ContractService) CreateContract(ctx context.Context, contract *models.Contract, creator string, channel string) (string, error) {
	fName := "service/contract/create"
	tracer := otel.Tracer("api")
	_, span := tracer.Start(ctx, fName)
	defer span.End()

	location, _ := time.LoadLocation("UTC")
	now, _ := time.Parse(time.RFC1123, time.Now().In(location).Format(time.RFC1123))

	col := service.bucket.DefaultCollection()
	contract.DocType = "contract"
	contract.Identifier = common.GenerateIdentifier(30)

	contract.Creator = creator
	contract.Channel = channel
	contract.Status = "active"
	contract.CreatedAt = now
	contract.LastUpdatedAt = now
	contract.LastUpdatedBy = contract.Creator
	_, err := col.Insert(contract_prefix+"/"+contract.Identifier, contract, nil)

	span.AddEvent("contract created")
	return contract.Identifier, err

}

func (service *ContractService) GetContract(ctx context.Context, contractId string) (any, error) {
	fName := "service/contract/get"
	tracer := otel.Tracer("api")
	_, span := tracer.Start(ctx, fName)
	defer span.End()

	col := service.bucket.DefaultCollection()
	doc, err := col.Get(contract_prefix+"/"+contractId, nil)
	if doc == nil {
		if err != nil {
			return "", errors.New("error: no contract found")
		}
	}
	return doc, err

}

func (service *ContractService) DeleteContract(ctx context.Context, contractId string, sessionedUser string) error {
	fName := "service/contract/delete"
	tracer := otel.Tracer("api")
	_, span := tracer.Start(ctx, fName)
	defer span.End()

	col := service.bucket.DefaultCollection()
	doc, err := col.Get(contract_prefix+"/"+contractId, nil)
	if doc == nil {
		if err != nil {
			return errors.New("error: no contract found")
		}
	}

	var wallet models.Contract
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
	_, err = col.Replace(contract_prefix+"/"+contractId, wallet, nil)
	return err

}

func (service *ContractService) MarkContractAsExpired(ctx context.Context, contractId string, sessionedUser string) error {
	fName := "service/contract/markContractAsExpired"
	tracer := otel.Tracer("api")
	_, span := tracer.Start(ctx, fName)
	defer span.End()

	col := service.bucket.DefaultCollection()
	doc, err := col.Get(contract_prefix+"/"+contractId, nil)
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

	wallet.Status = WALLET_STATUS_EXPIRED
	location, _ := time.LoadLocation("UTC")
	now, _ := time.Parse(time.RFC1123, time.Now().In(location).Format(time.RFC1123))
	wallet.LastUpdatedAt = now
	wallet.LastUpdatedBy = sessionedUser

	// TODO: need to convert it into soft delete
	_, err = col.Replace(contract_prefix+"/"+contractId, wallet, nil)
	return err

}

func (service *ContractService) Filter(ctx context.Context, queryString string, params map[string]interface{}, sortBy string, limit int) ([]*models.Contract, error) {
	fName := "service/contract/filter"
	tracer := otel.Tracer("api")
	_, span := tracer.Start(ctx, fName)
	defer span.End()

	// TODO: we can even make the retuning fields as input from calling methods insead of returning all fields
	query := "select data.* from `testbucket`.`_default`.`_default` data where type='contract' "
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

	return parseClusterRows(rows), nil
}

func parseClusterRows(rows *gocb.QueryResult) []*models.Contract {
	var contracts []*models.Contract
	for rows.Next() {
		var obj models.Contract
		err := rows.Row(&obj)
		if err != nil {
			panic(err)
		}
		contracts = append(contracts, &obj)
	}
	defer rows.Close()
	err := rows.Err()
	if err != nil {
		panic(err)
	}

	return contracts
}
