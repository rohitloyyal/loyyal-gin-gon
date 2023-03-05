package token

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.opencensus.io/trace"
)

func GenerateToken(email string, role string, name string, userIdentifier string) (string, error) {
	token_expiry, err := strconv.Atoi(os.Getenv("JWT_TOKEN_VALIDITY"))

	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["sub"] = email
	claims["name"] = name
	claims["aud"] = role
	claims["userIdentifier"] = userIdentifier

	location, _ := time.LoadLocation("UTC")
	claims["exp"] = time.Now().In(location).Add(time.Hour * time.Duration(token_expiry)).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))

}

func TokenValid(c *gin.Context) error {
	_, sp := trace.StartSpan(c, "api/v4/userSession")
	defer sp.End()

	tokenString := ExtractToken(c)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
	})
	if err != nil {
		return err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		username := fmt.Sprintf("%s", claims["sub"])
		role := fmt.Sprintf("%s", claims["aud"])

		sp.AddAttributes(
			trace.StringAttribute("token.username", username),
			trace.StringAttribute("token.role", role),
		)

		c.Header("X-API-USER-ID", username)
		c.Header("X-API-USER-Role", role)
		// TODO: need to make channel dynamic from the token
		c.Header("X-API-CHANNEL", "loyyalchannel")

		return nil
	}
	return nil
}

func ExtractToken(c *gin.Context) string {
	token := c.Query("token")
	if token != "" {
		return token
	}
	bearerToken := c.Request.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}

func ExtractTokenID(c *gin.Context) (string, error) {

	tokenString := ExtractToken(c)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		username := fmt.Sprintf("%.0f", claims["user_id"])

		return username, nil
	}
	return "", nil
}
