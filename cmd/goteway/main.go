package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/mac/goteway/internal/runtime"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := runtime.NewAppFromEnv()
	if err != nil {
		log.Fatalf("init app: %v", err)
	}

	if err := app.Run(ctx); err != nil {
		log.Fatalf("run app: %v", err)
	}
}
