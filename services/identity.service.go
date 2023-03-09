package services

import (
	"context"
	"errors"
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/utils/common"
	"github.com/loyyal/loyyal-be-contract/utils/token"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/crypto/bcrypt"
)

type IdentityService struct {
	cluster *gocb.Cluster
	bucket  *gocb.Bucket
}

const (
	identity_prefix = "user"
)

const (
	IDENTITY_STATUS_ACTIVE   = "active"
	IDENTITY_STATUS_DISABLED = "disabled"
)

func NewIdentity(cluster *gocb.Cluster, bucket *gocb.Bucket) IdentityService {
	return IdentityService{cluster: cluster, bucket: bucket}
}

func (service *IdentityService) CreateBootstrapIdentity(ctx context.Context, username string, password string) error {
	existingAdmin, err := service.Filter(ctx, "and username=$username and identityType=$identityType", map[string]interface{}{
		"username":     username,
		"identityType": "admin",
	}, "createdAt", -1)

	if len(existingAdmin) > 0 {
		return nil
	}

	var identity models.Identity
	identity.Username = username
	identity.Password = password
	identity.IdentityType = "admin"

	_, err = service.Create(ctx, &identity)
	return err
}

func (service *IdentityService) Login(ctx context.Context, user *models.Identity) (string, error) {
	fName := "service/identity/login"
	tracer := otel.Tracer("api")
	_, span := tracer.Start(ctx, fName)
	defer span.End()

	identities, err := service.Filter(ctx, "and username = $username and isDeleted=false", map[string]interface{}{
		"username": user.Username,
	}, "createdAt", 1)

	if err != nil {
		return "", err
	}
	if len(identities) != 1 || identities[0] == nil {
		return "", errors.New("error: no user found")
	}

	registeredUser := identities[0]
	err = VerifyPassword(user.Password, registeredUser.Password)
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return "", errors.New("error: invalid credentials")
	}

	if registeredUser.Status != "active" {
		return "", errors.New("error: this account is inactive or disabled")
	}

	// TODO: save session token in db
	// err = service.saveSession()
	// if err != nil {
	// 	return "", err
	// }

	// generate token
	token, err := token.GenerateToken(registeredUser.Username, registeredUser.IdentityType, registeredUser.PersonalDetails.FirstName, registeredUser.Identifier)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (service *IdentityService) Create(ctx context.Context, identity *models.Identity) (string, error) {
	fName := "service/identity/create"
	tracer := otel.Tracer("api")
	_, span := tracer.Start(ctx, fName)
	defer span.End()

	span.AddEvent("creating user identity")

	identity.DocType = "user"
	identity.Identifier = common.GenerateIdentifier(30)
	identity.Username = html.EscapeString(strings.TrimSpace(identity.Username))
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(identity.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	identity.Password = string(hashedPassword)
	identity.Channel = "loyyalchannel"
	if identity.IdentityType == "" {
		identity.IdentityType = "consumer"
	}
	if identity.Creator == "" {
		identity.Creator = "admin"
	}

	identity.Status = WALLET_STATUS_ACTIVE
	location, _ := time.LoadLocation("UTC")
	now, _ := time.Parse(time.RFC1123, time.Now().In(location).Format(time.RFC1123))

	identity.CreatedAt = now
	identity.LastUpdatedAt = now
	identity.LastUpdatedBy = identity.Creator

	col := service.bucket.DefaultCollection()
	_, err = col.Insert(identity_prefix+"/"+identity.Identifier, identity, nil)

	span.AddEvent("created user identity to couchbase")
	return identity.Identifier, err

}

func (service *IdentityService) Get(ctx context.Context, identityId string) (any, error) {
	fName := "service/identity/get"
	tracer := otel.Tracer("api")
	_, span := tracer.Start(ctx, fName)
	defer span.End()

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

	identity.Password = ""
	return identity, err
}

func (service *IdentityService) Update(ctx context.Context, identityId string, personalDetails models.PersonalDetails) error {
	fName := "service/identity/update"
	tracer := otel.Tracer("api")
	_, span := tracer.Start(ctx, fName)
	defer span.End()

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
	location, _ := time.LoadLocation("UTC")
	now, _ := time.Parse(time.RFC1123, time.Now().In(location).Format(time.RFC1123))

	identity.LastUpdatedAt = now
	identity.LastUpdatedBy = "admin"

	// TODO: need to convert it into soft delete
	_, err = col.Replace(identity_prefix+"/"+identityId, identity, nil)
	return err
}

func (service *IdentityService) Delete(ctx context.Context, identityId string, sessionedUser string) error {
	fName := "service/identity/delete"
	tracer := otel.Tracer("api")
	_, span := tracer.Start(ctx, fName)
	defer span.End()

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
	location, _ := time.LoadLocation("UTC")
	now, _ := time.Parse(time.RFC1123, time.Now().In(location).Format(time.RFC1123))
	identity.LastUpdatedAt = now
	identity.LastUpdatedBy = sessionedUser

	_, err = col.Replace(identity_prefix+"/"+identityId, identity, nil)
	span.AddEvent("identity updated")

	return err

}

func (service *IdentityService) Filter(ctx context.Context, queryString string, params map[string]interface{}, sortBy string, limit int) ([]*models.Identity, error) {
	fName := "service/identity/filter"
	tracer := otel.Tracer("api")
	_, span := tracer.Start(ctx, fName)
	defer span.End()

	// TODO: we can even make the retuning fields as input from calling methods insead of returning all fields
	query := "select data.* from `testbucket`.`_default`.`_default` data where type='user' "
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

	return parseIdentityRows(rows), nil
}

func parseIdentityRows(rows *gocb.QueryResult) []*models.Identity {
	var identities []*models.Identity
	for rows.Next() {
		var obj models.Identity
		err := rows.Row(&obj)
		if err != nil {
			panic(err)
		}
		obj.Password = ""
		identities = append(identities, &obj)
	}
	defer rows.Close()
	err := rows.Err()
	if err != nil {
		panic(err)
	}

	return identities
}
