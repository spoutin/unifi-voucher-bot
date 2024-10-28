package main

import (
	"context"
	"fmt"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"log/slog"
	"os"
)

var messageChannel MessageChannel

func InitSlack(channel MessageChannel, appToken, botToken string) *socketmode.Client {
	messageChannel = channel

	attrsInternal := []slog.Attr{{
		Key:   "source",
		Value: slog.AnyValue("slackInteral"),
	}}
	attrs := []slog.Attr{{
		Key:   "source",
		Value: slog.AnyValue("slack"),
	}}
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	internalSlackLogger := slog.NewLogLogger(handler.WithAttrs(attrsInternal), slog.LevelInfo)
	logger = slog.New(handler.WithAttrs(attrs))

	api := slack.New(
		botToken,
		slack.OptionDebug(false),
		slack.OptionLog(
			internalSlackLogger,
		),
		slack.OptionAppLevelToken(appToken),
	)

	client := socketmode.New(
		api,
		socketmode.OptionDebug(false),
		socketmode.OptionLog(
			internalSlackLogger,
		),
	)
	return client
}

func SocketMessageHandler(ctx context.Context, client *socketmode.Client) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-client.Events:
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				logger.Info("Connecting to Slack with Socket Mode...")
			case socketmode.EventTypeHello:
				logger.Info("Hello message received...")
			case socketmode.EventTypeConnectionError:
				logger.Error("Connection failed. Retrying later...")
				logger.Error(fmt.Sprintf("%v", evt))
			case socketmode.EventTypeConnected:
				logger.Info("Connected to Slack with Socket Mode.")
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					logger.Warn(fmt.Sprintf("Ignored %+v", evt))
					continue
				}
				logger.Debug(fmt.Sprintf("Event received: %+v", eventsAPIEvent))
				client.Ack(*evt.Request)

				EventApiReceived(client, eventsAPIEvent)
			case socketmode.EventTypeInteractive:
				callback, ok := evt.Data.(slack.InteractionCallback)
				if !ok {
					logger.Warn(fmt.Sprintf("Ignored %+v", evt))
					continue
				}
				logger.Debug(fmt.Sprintf("Interaction received: %+v", callback))

				EventInteractiveReceived(client, callback, evt)
			case socketmode.EventTypeSlashCommand:
				cmd, ok := evt.Data.(slack.SlashCommand)
				if !ok {
					logger.Warn(fmt.Sprintf("Ignored %+v", evt))
					continue
				}
				logger.Debug(fmt.Sprintf("Slash command received: %+v", cmd))

				EventSlashCommandReceived(client, evt)
			case socketmode.EventTypeDisconnect:
				logger.Info("Disconnected from Slack")
				return
			default:
				logger.Error(fmt.Sprintf("Unexpected event type received: %s", evt.Type))
			}
			break
		}
	}
}

func EventInteractiveReceived(client *socketmode.Client, callback slack.InteractionCallback, evt socketmode.Event) {
	logger.Warn(fmt.Sprintf("Interactive received %+v", callback))
}

func EventApiReceived(client *socketmode.Client, eventsAPIEvent slackevents.EventsAPIEvent) {
	logger.Debug(fmt.Sprintf("Event received: %+v", eventsAPIEvent))
}

func EventSlashCommandReceived(client *socketmode.Client, evt socketmode.Event) {
	logger.Warn(fmt.Sprintf("Slash command received: %+v", evt))

	p := func(message string) {
		payload := map[string]interface{}{
			"blocks": []slack.Block{
				slack.NewSectionBlock(
					&slack.TextBlockObject{
						Type: slack.MarkdownType,
						Text: fmt.Sprintf("```\n%s```", message),
					},
					nil,
					nil,
				),
			},
		}
		client.Ack(*evt.Request, payload)
	}

	messageChannel <- &Message{
		Process: p,
	}

}
