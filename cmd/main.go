package main

import (
	"app/internal/app"
	"context"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	app := app.New(ctx)

	go func() {
		app.SetupGracefulShutdown()
		cancel()
	}()

	app.Run()

}
