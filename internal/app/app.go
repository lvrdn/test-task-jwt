package app

import (
	"app/internal/config"
	"app/internal/handler"
	mockSender "app/internal/sender/mock"
	"app/internal/storage/postgresql"
	"app/pkg/logger"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
)

type app struct {
	server *http.Server
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

	app := &app{
		server: &http.Server{},
		wg:     &sync.WaitGroup{},
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
		cfg.AccessTokenKey,
		cfg.AccessTokenExpMinutes,
		cfg.RefreshTokenExpMinutes,
		storage,
		emailSender,
	)

	mux := http.NewServeMux()
	handler.SetRoutes(mux)

	app.server.Addr = ":" + cfg.HTTPport
	app.server.Handler = mux

	return app
}

func (a *app) Run() {

	const methodPtr string = "app.Run"

	msg := fmt.Sprintf("server started on %s", a.server.Addr)
	log.Println(msg)
	logger.Info(msg, methodPtr)

	a.server.ListenAndServe()
	a.wg.Wait()
}

func (a *app) SetupGracefulShutdown() {

	const methodPtr string = "app.SetupGracefulShutdown"

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	a.server.Shutdown(context.Background())

	msg := "server stopped"
	log.Println(msg)
	logger.Info(msg, methodPtr)
}
