package postgresql

import (
	"app/pkg/logger"
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type storageDBPostgreSQL struct {
	db *sql.DB
}

func NewStorage(ctx context.Context, wg *sync.WaitGroup, dbUsername, dbPassword, dbHost, dbName string) (*storageDBPostgreSQL, error) {

	const methodPtr string = "postgresql.NewStorage"

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		dbUsername,
		dbPassword,
		dbHost,
		dbName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Warn("open db connection failed", methodPtr, "error", err.Error())
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		logger.Warn("ping db failed", methodPtr, "error", err.Error())
		return nil, err
	}

	go func() {
		defer wg.Done()
		<-ctx.Done()
	}()

	return &storageDBPostgreSQL{
		db: db,
	}, nil
}

func (s *storageDBPostgreSQL) CheckGUID(guid string) (int, error) {
	var id int
	err := s.db.QueryRow("SELECT id FROM auth WHERE guid=$1", guid).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *storageDBPostgreSQL) AddNewRefreshToken(id int, hashedRefreshToken []byte, expDate time.Time) error {
	_, err := s.db.Exec("UPDATE auth SET hashed_refresh_token=$1, exp_date=$2 WHERE id=$3", hashedRefreshToken, expDate, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *storageDBPostgreSQL) GetRefreshTokenData(id int) ([]byte, time.Time, string, error) {
	var guid string
	var hashedRefreshToken []byte
	var expDate time.Time
	err := s.db.QueryRow("SELECT guid, hashed_refresh_token, exp_date FROM auth WHERE id=$1", id).Scan(&guid, &hashedRefreshToken, &expDate)
	if err != nil {
		return hashedRefreshToken, expDate, guid, err
	}
	return hashedRefreshToken, expDate, guid, nil
}
