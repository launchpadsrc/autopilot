package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"unicode/utf8"

	tele "gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/layout"

	"github.com/sashabaranov/go-openai"

	"launchpad.icu/autopilot/parsers"
)

type Bot struct {
	*tele.Bot

	ai *openai.Client

	parsers map[string]parsers.Parser
}

type Parsers = map[string]parsers.Parser

func New(ai *openai.Client, parsers Parsers) (*Bot, error) {
	lt, err := layout.New("bot.yml")
	if err != nil {
		return nil, err
	}

	b, err := tele.NewBot(lt.Settings())
	if err != nil {
		return nil, err
	}
	if err := b.SetCommands(lt.Commands()); err != nil {
		return nil, err
	}

	return &Bot{
		Bot:     b,
		ai:      ai,
		parsers: parsers,
	}, nil
}

func (b Bot) Start() {
	b.Handle("/keywords", b.onKeywords)
	b.Handle("/resume", b.onResume)
	b.Handle(tele.OnDocument, b.onResume)

	b.Bot.Start()
}

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

func (b Bot) sendHint(c tele.Context, hint string, v ...any) error {
	text := "💡 " + hint
	if len(v) > 0 {
		text += fmt.Sprintln(v...)
	}
	return c.Send(text)
}
