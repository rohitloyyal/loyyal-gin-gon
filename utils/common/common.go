package common

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

func GenerateIdentifier(length int) string {
	var chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	ll := len(chars)
	b := make([]byte, length)
	rand.Read(b) // generates len(b) random bytes
	for i := 0; i < length; i++ {
		b[i] = chars[int(b[i])%ll]
	}
	return string(b)
}

func NewRefID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}

func IsStructEmpty(object interface{}) bool {
	if object == nil {
		return true
	} else if object == "" {
		return true
	} else if object == false {
		return true
	}

	//Then see if it's a struct
	if reflect.ValueOf(object).Kind() == reflect.Struct {
		// and create an empty copy of the struct object to compare against
		empty := reflect.New(reflect.TypeOf(object)).Elem().Interface()
		if reflect.DeepEqual(object, empty) {
			return true
		}
	}
	return false
}

// randomID generates a random hex string in the UUID format:
// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
func randomID() (string, error) {
	var u [16]byte
	if _, err := rand.Read(u[:]); err != nil {
		return "", fmt.Errorf("genUUID: %v", err)
	}

	buf := make([]byte, 36)

	hex.Encode(buf[0:8], u[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], u[10:])

	return string(buf), nil
}

func PrepareCustomError(ctx *gin.Context, errorCode int, functionName string, displayMessage string, details string) {
	errID, err := randomID()
	if err != nil {
		errID = "unknown(" + err.Error() + ")"
	}

	ctx.Header("X-ErrID", errID)
	fmt.Printf("api error id=%q,path=%q,function name= %s,code=%d,err=%q,detail=%q", errID, ctx.Request.URL, functionName, errorCode, displayMessage, details)

	ctx.JSON(errorCode, gin.H{
		"code":    http.StatusBadRequest,
		"message": displayMessage,
		"error":   errID,
	})
	return
}

func PrepareCustomResponse(ctx *gin.Context, displayMessage string, body any) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": displayMessage,
		"body":    body,
	})
	return
}
