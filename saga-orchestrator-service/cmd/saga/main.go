package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
)

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	<-ctx.Done()
	fmt.Println("Saga Orchestrator Service stopped")
}
