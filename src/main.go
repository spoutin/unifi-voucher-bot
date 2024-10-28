package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {

	//Signal Catcher
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		wg.Add(1)
		defer wg.Done()
		sig := <-sigs
		logger.Warn(fmt.Sprintf("Received signal %s", sig))
		cancel()
	}()

	// Initial Configuration
	err := initConfig()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Start Voucher Service
	messageChannel := make(chan *Message)
	go func() {
		wg.Add(1)
		defer wg.Done()
		logger.Info("Starting Message Service")
		StartListener(ctx, messageChannel)
		logger.Info("Message Service Closed")
	}()

	// Start Slack Client
	client := InitSlack(messageChannel, SlackAppToken, SlackBotToken)
	go func() {
		wg.Add(1)
		defer wg.Done()
		logger.Info("Starting Slack Handler")
		SocketMessageHandler(ctx, client)
		logger.Info("Slack Handler Closed")
	}()

	go func() {
		wg.Add(1)
		defer wg.Done()
		logger.Info("Starting Slack Server")
		err := client.RunContext(ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error(err.Error())
			}
		}
		logger.Info("Slack Server stopped")
		cancel()
	}()

	// Waiting for signal to stop
	<-ctx.Done()

	// Ensure all process have closed.
	logger.Info("Waiting for all processes to terminate")
	wg.Wait()
	logger.Info("All process shut down")
}
