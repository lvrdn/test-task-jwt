package tokens

import "time"

type AccessToken struct {
	IssuedBy    string
	UserID      int
	IP          string
	MatchingKey string
	ExpDate     time.Time
	IatDate     time.Time
}

type RefreshToken struct {
	Value       string
	HashedValue []byte
	UserID      int
	MatchingKey string
	ExpDate     time.Time
}
