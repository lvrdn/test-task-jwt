package test

import (
	"app/internal/config"
	"app/internal/handler"
	mockSender "app/internal/sender/mock"
	"app/internal/storage/postgresql"
	"app/internal/tokens"
	"app/pkg/logger"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
)

type app struct {
	server *httptest.Server
	wg     *sync.WaitGroup
}

func New(ctx context.Context) *app {

	const methodPtr string = "app.New"

	err := logger.InitLogger()
	if err != nil {
		log.Fatalf("init logger failed: [%s]\n", err.Error())
	}

	cfg, err := config.GetConfig()
	if err != nil {
		log.Printf("get config failed: [%s]\n", err.Error())
		logger.Fatal("get config failed", methodPtr, "error", err.Error())
	}

	cfg.DBhost = "localhost"
	cfg.DBname = "test_db"
	cfg.DBusername = "root"
	cfg.DBpassword = "1234"
	cfg.MathcingKeyLen = 5
	cfg.RefreshTokenLen = 15
	cfg.AccessTokenExpMinutes = 30
	cfg.RefreshTokenExpMinutes = 43200
	cfg.AccessTokenKey = "!someKey321"

	app := &app{
		wg: &sync.WaitGroup{},
	}

	app.wg.Add(1)
	storage, err := postgresql.NewStorage(ctx, app.wg, cfg.DBusername, cfg.DBpassword, cfg.DBhost, cfg.DBname)
	if err != nil {
		log.Printf("get storage failed: [%s]\n", err.Error())
		logger.Fatal("get storage failed", methodPtr, "error", err.Error())
	}

	app.wg.Add(1)
	emailSender, err := mockSender.NewMockEmailSender(ctx, app.wg)
	if err != nil {
		log.Printf("get email sender failed: [%s]\n", err.Error())
		logger.Fatal("get email sender failed", methodPtr, "error", err.Error())
	}

	handler := handler.NewHandler(
		storage,
		emailSender,
		tokens.NewTokenManager(cfg.RefreshTokenLen, cfg.MathcingKeyLen, cfg.AccessTokenKey, cfg.AccessTokenExpMinutes, cfg.RefreshTokenExpMinutes),
	)

	mux := http.NewServeMux()
	handler.SetRoutes(mux)

	app.server = httptest.NewServer(mux)
	log.Println("test server started")

	return app
}

func (a *app) WaitClosingProcceses() {
	a.wg.Wait()
}
