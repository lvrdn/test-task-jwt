package tokens

import "time"

type TokenManager interface {
	CreateRefreshToken(id int) (*RefreshToken, error)
	CreateSignedAccessToken(id int, addr, matchingKey string) (string, error)
	ParseRefreshTokenFromReq(refreshTokenFromReq string) (*RefreshToken, error)
	ParseAccessTokenFromReq(accessTokenFromReq string) (*AccessToken, error)
	GetAccessTokenExpTime() time.Duration
	GetRefreshTokenExpTime() time.Duration
	CompareHashAndToken(hash []byte, token string) bool
}
