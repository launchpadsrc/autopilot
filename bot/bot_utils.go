package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"time"
	"unicode/utf8"

	tele "gopkg.in/telebot.v4"
)

// SendJSON sends a JSON indented repr of the provided value.
// If the resulting string is too long, it sends it as a file attachment instead.
func (b Bot) SendJSON(c tele.Context, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	jsonstr := "```json\n" + string(data) + "\n```"
	if utf8.RuneCountInString(jsonstr) <= 4096 {
		return c.Send(jsonstr, tele.ModeMarkdownV2)
	}

	go c.Notify(tele.UploadingDocument)

	return c.Send(&tele.Document{
		File:     tele.FromReader(bytes.NewReader(data)),
		FileName: reflect.TypeOf(v).String() + ".json",
	})
}

// WithNotify continuously sends a chat action to the user until cancelled.
func (b Bot) WithNotify(c tele.Context, action tele.ChatAction) func() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				c.Notify(action)
				<-time.After(2 * time.Second)
			}
		}
	}()
	return cancel
}

// AfterFunc runs the provided function after the specified duration.
func (b Bot) AfterFunc(d time.Duration, f func() error) error {
	<-time.After(d)
	return f()
}
