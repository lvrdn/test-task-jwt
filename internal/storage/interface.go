package storage

import "time"

type Storage interface {
	CheckGUID(guid string) (int, error)
	AddNewRefreshToken(id int, hashedRefreshToken []byte, expDate time.Time) error
	GetRefreshTokenData(id int) ([]byte, time.Time, string, error)
}
