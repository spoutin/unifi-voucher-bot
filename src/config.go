package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

var UnifiBaseURL string
var UnifiSSLVerify bool
var UnifiUser string
var UnifiPassword string
var SlackBotToken string
var SlackAppToken string

var logger *slog.Logger

func initConfig() error {

	// Logging
	logHandler := slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{Level: slog.LevelInfo},
	)
	logger = slog.New(logHandler).With("source", "main")
	slog.SetDefault(logger)

	// Unifi
	errorFlag := false
	var err error

	UnifiBaseURL = os.Getenv("UNIFI_BASE_URL")
	if UnifiBaseURL == "" {
		logger.Error("UNIFI_BASE_URL environment variable not set")
		errorFlag = true
	}

	UnifiUser = os.Getenv("UNIFI_USER")
	if UnifiUser == "" {
		logger.Error("UNIFI_USER environment variable not set")
		errorFlag = true
	}
	UnifiPassword = os.Getenv("UNIFI_PASSWORD")
	if UnifiPassword == "" {
		logger.Error("UNIFI_PASSWORD environment variable not set")
		errorFlag = true
	}

	unifiSSLVerify := os.Getenv("UNIFI_SSL_VERIFY")
	if unifiSSLVerify == "" {
		UnifiSSLVerify = true
	} else {
		UnifiSSLVerify, err = strconv.ParseBool(unifiSSLVerify)
		if err != nil {
			logger.Error("UNIFI_SSL_VERIFY environment must be a boolean value")
			errorFlag = true
		}
	}

	// Slack
	SlackAppToken = os.Getenv("SLACK_APP_TOKEN")
	if SlackAppToken == "" {
		logger.Error("SLACK_APP_TOKEN must be set.")
		errorFlag = true
	}

	if !strings.HasPrefix(SlackAppToken, "xapp-") {
		logger.Error("SLACK_APP_TOKEN must have the prefix \"xapp-\".")
		errorFlag = true
	}

	SlackBotToken = os.Getenv("SLACK_BOT_TOKEN")
	if SlackBotToken == "" {
		logger.Error("SLACK_BOT_TOKEN must be set.")
		errorFlag = true
	}

	if !strings.HasPrefix(SlackBotToken, "xoxb-") {
		logger.Error("SLACK_BOT_TOKEN must have the prefix \"xoxb-\".")
		errorFlag = true
	}

	if errorFlag {
		return fmt.Errorf("invalid environment variable(s)")
	}
	return nil

}
