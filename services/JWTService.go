package services

import (
	"net/http"
	"strings"
)

func GenerateToken() {

}

func VerifyToken() {
	// get bearer token from header

	// check if valid session is stored in database
	// verify, signed by,timestamp, expiry

	// add response header

	// c.Header()
	// r.Header.Add(UserIDKey, session.User)
	// r.Header.Add(ChannelKey, session.Channel)
	// r.Header.Add(UserRoleKey, session.Role)
	// r.Header.Add(DomainKey, r.Host)
}

// bearerAuth parses the Authorization header for a Bearer token.
//
// example curl:
//
//	curl -H "Authorization: Bearer <ACCESS_TOKEN>" https://loyyal.loyyalbeta.com/v4/identity -d {}
func bearerAuth(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return ""
	}
	return auth[len(prefix):]
}

// func (k *Key) ParseToken(token string) (*Token, error) {
// 	c := jwt.StandardClaims{}
// 	jt, err := jwt.ParseWithClaims(token, &c, func(t *jwt.Token) (interface{}, error) {
// 		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
// 			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
// 		}
// 		return k.Bytes, nil
// 	})
// 	if err != nil || (c.Id == "" && c.IssuedAt == 0) {
// 		return k.ParseLegacyToken(token)
// 	}
// 	return &Token{
// 		Secret:   token,
// 		Token:    jt,
// 		CUID:     c.Id,
// 		Issued:   c.IssuedAt,
// 		Subject:  c.Subject,
// 		Issuer:   c.Issuer,
// 		Audience: c.Audience,
// 	}, nil
// }
