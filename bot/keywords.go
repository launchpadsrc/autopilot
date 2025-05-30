package bot

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v4"

	"launchpad.icu/autopilot/core/gapanalysis"
)

func (b Bot) onKeywords(c tele.Context) error {
	k, _ := strconv.Atoi(c.Message().Payload)
	if k == 0 {
		k = 5 // default number of keywords to extract
	}

	links := strings.Split(c.Text(), "\n")[1:]
	if len(links) == 0 {
		return b.sendHint(c, "No links provided.")
	}

	var jds []string
	for _, link := range links {
		uri, err := url.Parse(link)
		if err != nil {
			return b.sendHint(c, "Invalid link:", link)
		}

		parser, ok := b.parsers[uri.Hostname()]
		if !ok {
			return b.sendHint(c, "No parser for:", uri.Hostname())
		}

		go c.Notify(tele.Typing)

		job, err := parser.ParseJob(link)
		if err != nil {
			return b.sendHint(c, "Failed to parse job:", err)
		}

		jds = append(jds, job.Description)
	}

	keywords, err := gapanalysis.NewKeywordsExtractor(b.ai).Extract(k, jds)
	if err != nil {
		return b.sendHint(c, "Failed to extract keywords:", err)
	}

	data, _ := json.MarshalIndent(keywords, "", "  ")
	return c.Send("```json\n"+string(data)+"```", tele.ModeMarkdownV2)
}
