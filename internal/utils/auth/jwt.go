package auth

import (
	"golang-starter/internal/config"
	"golang-starter/internal/db"
	"golang-starter/internal/utils/encryption"
	"log"
	"time"

	"github.com/golang-jwt/jwt"
)

type TokenDTO struct {
	Type         string `json:"type"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshDTO struct {
	RefreshToken string `json:"refresh_token"`
	Expired      int64  `json:"expired"`
}

// Sign ins method to generate jwt token and refresh token
// it has ... parameter
// userdata is map data, it's using for passing user data
// default expired time is 60 second
func Sign(claims jwt.MapClaims) TokenDTO {
	timeNow := time.Now()
	tokenExpired := timeNow.Add(time.Second * config.Get().JwtTokenExpired).Unix()

	if claims["id"] == nil {
		return TokenDTO{}
	}

	token := jwt.New(jwt.SigningMethodRS256)
	// setup userdata
	var _, checkExp = claims["exp"]
	var _, checkIat = claims["exp"]

	// if user didn't define claims expired
	if checkExp == false {
		claims["exp"] = tokenExpired
	}
	// if user didn't define claims iat
	if checkIat == false {
		claims["iat"] = timeNow.Unix()
	}
	claims["token_type"] = "access_token"

	token.Claims = claims

	authToken := new(TokenDTO)
	tokenString, err := token.SignedString(config.Get().PrivateKey)
	if err != nil {
		log.Fatalln(err)
		return TokenDTO{}
	}

	authToken.Token = tokenString
	authToken.Type = config.Get().JwtTokenType

	//create refresh token
	refreshToken := jwt.New(jwt.SigningMethodRS256)
	refreshTokenExpired := timeNow.Add(time.Second * config.Get().JwtRefreshExpired).Unix()

	claims["exp"] = refreshTokenExpired
	claims["token_type"] = "refresh_token"
	refreshToken.Claims = claims

	refreshTokenString, err := refreshToken.SignedString(config.Get().PrivateKey)

	if err != nil {
		return TokenDTO{}
	}
	authToken.RefreshToken = refreshTokenString

	//save token to local db
	go func() {
		encryptionRefreshToken := encryption.AesCFBEncryption(refreshTokenString, config.Get().AppKey)
		scribleDB := db.NewScribleClient()
		scribleDB.Query().Write("refresh_token", claims["id"].(string), RefreshDTO{RefreshToken: encryptionRefreshToken, Expired: refreshTokenExpired})
	}()

	return TokenDTO{
		Type:         "Bearer",
		Token:        authToken.Token,
		RefreshToken: authToken.RefreshToken,
	}
}
