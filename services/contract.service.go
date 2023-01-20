package services

import (
	"context"

	"github.com/couchbase/gocb/v2"
	"github.com/loyyal/loyyal-be-contract/models"
)

type ContractService struct {
	bucket *gocb.Bucket
	ctx    context.Context
}

func New(bucket *gocb.Bucket, ctx context.Context) ContractService {
	return ContractService{bucket: bucket, ctx: ctx}
}

func (service *ContractService) Create(contract *models.Contract) error {
	col := service.bucket.DefaultCollection()
	_, err := col.Upsert(contract.ContractId, contract, nil)

	return err

}

func (service *ContractService) Get(contractId string) (any, error) {
	col := service.bucket.DefaultCollection()
	doc, err := col.Get(contractId, nil)


	return doc, err

}

func (service *ContractService) Filter(contract *models.Contract) error {
	col := service.bucket.DefaultCollection()
	_, err := col.Upsert(contract.ContractId, contract, nil)

	return err

}
