package oauth2

import (
	"context"

	"google.golang.org/api/idtoken"
)

// https://developers.google.com/identity/sign-in/web/backend-auth?authuser=6#verify-the-integrity-of-the-id-token

// func verifyIdToken(idToken string) (*oauth2.Tokeninfo, error) {
// 	// https://pkg.go.dev/google.golang.org/api@v0.49.0/oauth2/v2#hdr-Creating_a_client
// 	ctx := context.Background()
// 	oauth2Service, err := oauth2.NewService(ctx, option.WithAPIKey(conf.GoogleSerivceAPIKey))
// 	if err != nil {
// 		return nil, err
// 	}
// 	tokenInfoCall := oauth2Service.Tokeninfo()
// 	tokenInfoCall.IdToken(idToken)
// 	if tokenInfo, err := tokenInfoCall.Do(); err != nil {
// 		return nil, err
// 	} else {
// 		return tokenInfo, nil
// 	}
// }

// https://stackoverflow.com/questions/36716117/validating-google-sign-in-id-token-in-go
// https://developers.google.com/identity/sign-in/web/backend-auth#using-a-google-api-client-library
func VerifyIdToken(idToken, googleOauth2WebClientID string) (email, name, firstName, lastName string, err error) {
	var token string // this comes from your web or mobile app maybe
	// const googleClientId ""=  // from credentials in the Google dev console

	// tokenValidator, err := idtoken.NewValidator(context.Background(), option.WithAPIKey(conf.GoogleSerivceAPIKey))
	tokenValidator, err := idtoken.NewValidator(context.Background())

	if err != nil {
		return "", "", "", "", err
	}

	payload, err := tokenValidator.Validate(context.Background(), token, googleOauth2WebClientID)
	if err != nil {
		return "", "", "", "", err
	}
	// "email": "testuser@gmail.com",
	//  "email_verified": "true",
	//  "name" : "Test User",
	//  "picture": "https://lh4.googleusercontent.com/-kYgzyAWpZzJ/ABCDEFGHI/AAAJKLMNOP/tIXL9Ir44LE/s99-c/photo.jpg",
	//  "given_name": "Test",
	//  "family_name": "User",
	//  "locale": "en"
	email = payload.Claims["email"].(string)
	name = payload.Claims["name"].(string)
	firstName = payload.Claims["given_name"].(string)
	lastName = payload.Claims["family_name"].(string)
	return email, name, firstName, lastName, nil
}
