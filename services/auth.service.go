package services

import (
	"context"
	"html"
	"strings"

	"github.com/couchbase/gocb/v2"
	"github.com/loyyal/loyyal-be-contract/models"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	cluster *gocb.Cluster
	bucket  *gocb.Bucket
	ctx     context.Context
}

func NewAuthService(cluster *gocb.Cluster, bucket *gocb.Bucket, ctx context.Context) AuthService {
	return AuthService{cluster: cluster, bucket: bucket, ctx: ctx}
}

func (service *AuthService) Register(user *models.User) error {

	user.DocType = "user"
	user.Username = html.EscapeString(strings.TrimSpace(user.Username))
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	col := service.bucket.DefaultCollection()
	_, err = col.Insert(user.Username, user, nil)

	return err

}

func (service *AuthService) saveSession() error {
	return nil
}

func VerifyPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
