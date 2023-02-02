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

type ContractService struct {
	cluster *gocb.Cluster
	bucket  *gocb.Bucket
	ctx     context.Context
}

const (
	contract_prefix = "contract"
)

func NewContract(cluster *gocb.Cluster, bucket *gocb.Bucket, ctx context.Context) ContractService {
	return ContractService{cluster: cluster, bucket: bucket, ctx: ctx}
}

func (service *ContractService) CreateContract(contract *models.Contract, creator string, channel string) (string, error) {
	col := service.bucket.DefaultCollection()
	contract.Identifier = common.GenerateIdentifier(30)
	contract.Creator = creator
	contract.Channel = channel
	contract.Status = "active"
	contract.CreatedAt = time.Now()
	contract.LastUpdatedAt = time.Now()
	contract.LastUpdatedBy = contract.Creator
	_, err := col.Insert(contract_prefix+"/"+contract.Identifier, contract, nil)

	return contract.Identifier, err

}

func (service *ContractService) GetContract(contractId string) (any, error) {
	col := service.bucket.DefaultCollection()
	doc, err := col.Get(contract_prefix+"/"+contractId, nil)
	if doc == nil {
		if err != nil {
			return "", errors.New("error: no contract found")
		}
	}
	return doc, err

}

func (service *ContractService) DeleteContract(contractId string, sessionedUser string) error {
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

	wallet.IsDeleted = true
	wallet.LastUpdatedAt = time.Now()
	wallet.LastUpdatedBy = sessionedUser

	// TODO: need to convert it into soft delete
	_, err = col.Replace(contract_prefix+"/"+contractId, wallet, nil)
	return err

}

func (service *ContractService) Filter(queryString string, params map[string]interface{}, sortBy string, limit int) ([]*models.Contract, error) {

	// TODO: we can even make the retuning fields as input from calling methods insead of returning all fields
	query := "select data.* from `testbucket`.`_default`.`_default` data where type='contract' "
	query += queryString
	query += " order by " + sortBy
	query += " limit " + strconv.Itoa(limit)

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
