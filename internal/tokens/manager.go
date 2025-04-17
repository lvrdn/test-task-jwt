package tokens

import (
	"fmt"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type tokenManager struct {
	refreshTokenLen        int
	matchingKeyLen         int
	accessTokenKey         string
	accessTokenExpMinutes  time.Duration
	refreshTokenExpMinutes time.Duration
}

func NewTokenManager(refreshTokenLen, matchingKeyLen int, accessTokenKey string, accessTokenExpMinutes, refreshTokenExpMinutes int) *tokenManager {
	return &tokenManager{
		refreshTokenLen:        refreshTokenLen,
		matchingKeyLen:         matchingKeyLen,
		accessTokenKey:         accessTokenKey,
		accessTokenExpMinutes:  time.Duration(accessTokenExpMinutes),
		refreshTokenExpMinutes: time.Duration(refreshTokenExpMinutes),
	}
}

func (tm *tokenManager) GetAccessTokenExpTime() time.Duration {
	return tm.accessTokenExpMinutes
}

func (tm *tokenManager) GetRefreshTokenExpTime() time.Duration {
	return tm.refreshTokenExpMinutes
}

func (tm *tokenManager) CreateRefreshToken(userID int) (*RefreshToken, error) {
	matchingKey := randomString(tm.matchingKeyLen)
	refreshTokenValue := fmt.Sprintf("%s.%s", randomString(tm.refreshTokenLen), matchingKey)
	hashedValue, err := bcrypt.GenerateFromPassword([]byte(refreshTokenValue), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &RefreshToken{
		Value:       refreshTokenValue,
		HashedValue: hashedValue,
		UserID:      userID,
		MatchingKey: matchingKey,
		ExpDate:     time.Now().Add(time.Minute * tm.refreshTokenExpMinutes),
	}, nil
}

func (tm *tokenManager) CreateSignedAccessToken(id int, ip, matchingKey string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":          "authApp",
		"user_id":      id,
		"ip":           ip,
		"matching_key": matchingKey,
		"exp":          jwt.NewNumericDate(now.Add(time.Duration(tm.accessTokenExpMinutes) * time.Minute)),
		"iat":          jwt.NewNumericDate(now),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	signedAccessToken, err := accessToken.SignedString([]byte(tm.accessTokenKey))
	if err != nil {
		return "", err
	}

	return signedAccessToken, nil
}

func (tm *tokenManager) ParseRefreshTokenFromReq(refreshTokenFromReq string) (*RefreshToken, error) {
	values := strings.Split(refreshTokenFromReq, ".")
	if len(values) != 2 {
		return nil, fmt.Errorf("refresh token must have value and matching key")
	}
	return &RefreshToken{
		Value:       refreshTokenFromReq,
		MatchingKey: values[1],
	}, nil
}

func (tm *tokenManager) ParseAccessTokenFromReq(accessTokenFromReq string) (*AccessToken, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(
		accessTokenFromReq,
		&claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(tm.accessTokenKey), nil
		},
	)
	if err != nil {
		return nil, err
	}

	issuedBy, ok := claims["iss"].(string)
	if !ok {
		return nil, fmt.Errorf("no issued by in token")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("no user id in token")
	}

	ip, ok := claims["ip"].(string)
	if !ok {
		return nil, fmt.Errorf("no ip in token")
	}

	matchingKey, ok := claims["matching_key"].(string)
	if !ok {
		return nil, fmt.Errorf("no matching key in token")
	}

	expDateFloat, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("no exp date in token")
	}
	expDate := time.Unix(int64(expDateFloat), 0)

	iatDateFloat, ok := claims["iat"].(float64)
	if !ok {
		return nil, fmt.Errorf("no iat date in token")
	}
	iatDate := time.Unix(int64(iatDateFloat), 0)

	return &AccessToken{
		IssuedBy:    issuedBy,
		UserID:      int(userID),
		IP:          ip,
		MatchingKey: matchingKey,
		ExpDate:     expDate,
		IatDate:     iatDate,
	}, nil
}

func (tm *tokenManager) CompareHashAndToken(hash []byte, token string) bool {
	err := bcrypt.CompareHashAndPassword(hash, []byte(token))
	if err != nil {
		return false
	}
	return true
}
