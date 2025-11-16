package jwt

import (
	"fmt"
	"testing"
	"time"

	"github.com/adamesong/go-util/random"
	// "github.com/dgrijalva/jwt-go"
	// "github.com/golang-jwt/jwt/v4"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestToken(t *testing.T) {
	claim := CustomClaims{
		UserID: uuid.NewString(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    random.RandomString(5),
			Subject:   random.RandomString(5),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 5)),
		},
	}
	jwtSecret := random.RandomString(12)
	jwtSign := NewJwtSign(jwtSecret)
	token, err := jwtSign.CreateToken(claim)

	assertion := assert.New(t)
	assertion.Nil(err)
	assertion.NotEqual("", token)
	fmt.Println("token: ", token)

	parsedClaim, pErr := jwtSign.ParseToken(token)
	assertion.Nil(pErr)
	assertion.Equal(parsedClaim.Issuer, claim.Issuer)
	assertion.Equal(parsedClaim.IssuedAt, claim.IssuedAt)
	assertion.Equal(parsedClaim.Subject, claim.Subject)
	assertion.Equal(parsedClaim.ExpiresAt, claim.ExpiresAt)
	assertion.Equal(parsedClaim.UserID, claim.UserID)
	fmt.Println("parsed claim: ", parsedClaim)
}

func TestParseToken(t *testing.T) {

}
