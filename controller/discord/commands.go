package discord

import (
	"github.com/bwmarrin/discordgo"
	"regexp"
)

func (dcc *Client) RegisterDiscordCommand(cmd string, callback func(msg *discordgo.MessageCreate, args string)) {
	rsp := regexp.MustCompile("^(?i:" + cmd + ")(?:\\s+(.*))?$")
	dcc.cleanup.Add(1)
	go func() {
		tap, done := dcc.TapChat()
		defer done()
		for {
			select {
			case <-dcc.Context.Done():
				dcc.cleanup.Done()
				return
			case msg := <-tap:
				parts := rsp.FindStringSubmatch(msg.Content)
				if parts != nil {
					args := parts[1]
					go callback(msg, args)
				}
			}
		}
	}()
}
