package services

import (
	"context"
	"errors"
	"html"
	"strings"

	"github.com/couchbase/gocb/v2"
	"github.com/loyyal/loyyal-be-contract/models"
	"github.com/loyyal/loyyal-be-contract/utils/token"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	bucket *gocb.Bucket
	ctx    context.Context
}

func NewUserService(bucket *gocb.Bucket, ctx context.Context) UserService {
	return UserService{bucket: bucket, ctx: ctx}
}

func (service *UserService) Login(user *models.User) (string, error) {
	// get user details and check for hashed password
	col := service.bucket.DefaultCollection()
	doc, err := col.Get(user.Username, nil)
	if doc == nil {
		if err != nil {
			return "", errors.New("error: no user found")
		}
	}

	var registeredUser models.User
	err = doc.Content(&registeredUser)
	if err != nil {
		return "", err
	}

	err = VerifyPassword(user.Password, registeredUser.Password)
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return "", errors.New("error: invalid credentials")
	}
	// save session token in db
	err = service.saveSession()
	if err != nil {
		return "", err
	}

	// generate token
	token, err := token.GenerateToken(user.Username)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (service *UserService) Register(user *models.User) error {

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

func (service *UserService) saveSession() error {
	return nil
}

func VerifyPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
