package zapslack

import (
	"errors"
	"sync"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/bluele/slack"
)

// SlackHook is a zap Hook for dispatching messages to the specified
// channel on Slack.
type SlackHook struct {
	// Messages with a log level not contained in this array
	// will not be dispatched. If nil, all messages will be dispatched.
	AcceptedLevels []zapcore.Level
	HookURL        string // Webhook URL

	// slack post parameters
	Username  string // display name
	Channel   string // `#channel-name`
	IconEmoji string // emoji string ex) ":ghost:":
	IconURL   string // icon url

	FieldHeader string        // a header above field data
	Timeout     time.Duration // request timeout
	Async       bool          // if async is true, send a message asynchronously.

	hook *slack.WebHook

	once sync.Once
}

func NewSlackHook(hookURL string, level zapcore.Level) *SlackHook {
	return &SlackHook{
		HookURL:        hookURL,
		AcceptedLevels: []zapcore.Level{level},
	}
}

func (sh *SlackHook) GetHook() func(zapcore.Entry) error {
	return func(e zapcore.Entry) error {
		sh.once.Do(func() {
			sh.hook = slack.NewWebHook(sh.HookURL)
		})
		if !sh.isAcceptedLevel(e.Level) {
			return nil
		}
		payload := &slack.WebHookPostPayload{
			Username:  sh.Username,
			Channel:   sh.Channel,
			IconEmoji: sh.IconEmoji,
			IconUrl:   sh.IconURL,
		}
		color, _ := LevelColorMap[e.Level]

		attachment := slack.Attachment{}
		payload.Attachments = []*slack.Attachment{&attachment}
		attachment.Text = e.Message
		attachment.Fallback = e.Message
		attachment.Color = color

		if sh.Async {
			go sh.postMessage(payload)
			return nil
		}

		return sh.postMessage(payload)
	}
}

func (sh *SlackHook) postMessage(payload *slack.WebHookPostPayload) error {
	if sh.Timeout <= 0 {
		return sh.hook.PostMessage(payload)
	}

	ech := make(chan error, 1)
	go func(ch chan error) {
		ch <- nil
		ch <- sh.hook.PostMessage(payload)
	}(ech)
	<-ech

	select {
	case err := <-ech:
		return err
	case <-time.After(sh.Timeout):
		return TimeoutError
	}
}

// Levels sets which levels to sent to slack
func (sh *SlackHook) Levels() []zapcore.Level {
	if sh.AcceptedLevels == nil {
		return AllLevels
	}
	return sh.AcceptedLevels
}

func (sh *SlackHook) isAcceptedLevel(level zapcore.Level) bool {
	for _, lv := range sh.Levels() {
		if lv == level {
			return true
		}
	}
	return false
}

// Supported log levels
var AllLevels = []zapcore.Level{
	zapcore.DebugLevel,
	zapcore.InfoLevel,
	zapcore.WarnLevel,
	zapcore.ErrorLevel,
	zapcore.FatalLevel,
	zapcore.PanicLevel,
}

var LevelColorMap = map[zapcore.Level]string{
	zapcore.DebugLevel: "#9B30FF",
	zapcore.InfoLevel:  "good",
	zapcore.WarnLevel:  "warning",
	zapcore.ErrorLevel: "danger",
	zapcore.FatalLevel: "danger",
	zapcore.PanicLevel: "danger",
}

var TimeoutError = errors.New("Request timed out")

// LevelThreshold - Returns every logging level above and including the given parameter.
func LevelThreshold(l zapcore.Level) []zapcore.Level {
	for i := range AllLevels {
		if AllLevels[i] == l {
			return AllLevels[i:]
		}
	}
	return []zapcore.Level{}
}
