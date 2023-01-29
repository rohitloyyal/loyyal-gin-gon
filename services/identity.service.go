package services

import (
	"context"
	"errors"
	"html"
	"strings"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/utils/common"
	"golang.org/x/crypto/bcrypt"
)

type IdentityService struct {
	bucket *gocb.Bucket
	ctx    context.Context
}

const (
	identity_prefix = "user"
)

func NewIdentity(bucket *gocb.Bucket, ctx context.Context) IdentityService {
	return IdentityService{bucket: bucket, ctx: ctx}
}

func (service *IdentityService) Create(identity *models.Identity) (string, error) {
	identity.DocType = "user"
	identity.Identifier = common.GenerateIdentifier(30)
	identity.Username = html.EscapeString(strings.TrimSpace(identity.Username))
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(identity.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	identity.Password = string(hashedPassword)
	identity.Channel = "loyyalchannel"
	identity.IdentityType = "consumer"
	identity.Creator = "admin"
	identity.Status = "active"

	identity.CreatedAt = time.Now()
	identity.LastUpdatedAt = time.Now()
	identity.LastUpdatedBy = identity.Creator

	col := service.bucket.DefaultCollection()
	_, err = col.Insert(identity_prefix+"/"+identity.Identifier, identity, nil)

	return identity.Identifier, err

}

func (service *IdentityService) Get(identityId string) (any, error) {
	col := service.bucket.DefaultCollection()
	doc, err := col.Get(identity_prefix+"/"+identityId, nil)
	if doc == nil {
		if err != nil {
			return nil, errors.New("error: no user found")
		}
	}

	var identity models.Identity
	doc.Content(&identity)
	if err != nil {
		return nil, err
	}

	return identity, err
}

func (service *IdentityService) Filter(contract *models.Identity) error {
	col := service.bucket.DefaultCollection()
	_, err := col.Upsert(contract.Identifier, contract, nil)

	return err

}

func (service *IdentityService) Update(identityId string, personalDetails models.PersonalDetails) error {
	col := service.bucket.DefaultCollection()
	doc, err := col.Get(identity_prefix+"/"+identityId, nil)
	if doc == nil {
		if err != nil {
			return errors.New("error: no user found")
		}
	}

	var identity models.Identity
	err = doc.Content(&identity)
	if err != nil {
		return err
	}

	identity.PersonalDetails = personalDetails
	identity.LastUpdatedAt = time.Now()
	identity.LastUpdatedBy = "admin"

	// TODO: need to convert it into soft delete
	_, err = col.Replace(identity_prefix+"/"+identityId, identity, nil)
	return err
}

func (service *IdentityService) Delete(identityId string, sessionedUser string) error {
	col := service.bucket.DefaultCollection()
	doc, err := col.Get(identity_prefix+"/"+identityId, nil)
	if doc == nil {
		if err != nil {
			return errors.New("error: no identity found")
		}
	}

	var identity models.Identity
	err = doc.Content(&identity)
	if err != nil {
		return err
	}

	identity.IsDeleted = true
	identity.LastUpdatedAt = time.Now()
	identity.LastUpdatedBy = sessionedUser

	_, err = col.Replace(identity_prefix+"/"+identityId, identity, nil)
	// _, err := col.Remove(identity_prefix+""+identityId, nil)

	return err

}
