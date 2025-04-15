package postgresql

import (
	"context"
	"sync"
)

type storageDBPostgreSQL struct {
}

func NewStorage(ctx context.Context, wg *sync.WaitGroup, dbUsername, dbPassword, dbHost, dbName string) (*storageDBPostgreSQL, error) {
	// dsn := fmt.Sprintf(
	// 	"postgres://%s:%s@%s/%s?sslmode=disable",
	// 	dbUsername,
	// 	dbPassword,
	// 	dbHost,
	// 	dbName,
	// )

	go func() {
		defer wg.Done()
		<-ctx.Done()
	}()

	return &storageDBPostgreSQL{}, nil
}
