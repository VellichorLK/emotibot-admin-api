package emotijwt

import (
	"fmt"
	"time"

	"emotibot.com/emotigo/pkg/logger"
	jwt "github.com/dgrijalva/jwt-go"
)

const (
	signSecret = "keyForTest"
	authApp    = "simple-auth"
)

var jwtExpireSeconds = 60 * 60

type CustomClaims struct {
	Custom interface{} `json:"custom"`
	jwt.StandardClaims
}

func SetJWTExpireTime(seconds int) {
	jwtExpireSeconds = seconds
}

func GetJWTTokenWithCustomInfo(custom interface{}) (string, error) {
	now := time.Now()
	expireSecond := jwtExpireSeconds // token will expired after 1 hr

	// Create the Claims
	customClaims := &CustomClaims{
		custom,
		jwt.StandardClaims{
			Issuer:    authApp,
			ExpiresAt: now.Unix() + int64(expireSecond),
			NotBefore: now.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims)
	ss, err := token.SignedString([]byte(signSecret))
	return ss, err
}

func ResolveJWTToken(ss string) (interface{}, error) {
	token, err := jwt.Parse(ss, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(signSecret), nil
	})

	if err != nil {
		logger.Error.Printf("Resolve fail: %s", err.Error())
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["custom"], nil
	}
	return nil, err
}
