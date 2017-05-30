package main

import (
	"github.com/bluele/zapslack"
	"go.uber.org/zap"
)

var (
	// Please rewrite it with your webhook URL
	slackWebHookURL = "https://hooks.slack.com/services/XXXXX/YYYYY/ZZZZZ"
)

func main() {
	logger, _ := zap.NewProduction()

	// Send a notification to slack at only error, fatal, panic level
	logger = logger.WithOptions(
		zap.Hooks(zapslack.NewSlackHook(slackWebHookURL, zap.ErrorLevel).GetHook()),
	)

	logger.Debug("don't need to send a message")
	logger.Error("an error happened!")
}
