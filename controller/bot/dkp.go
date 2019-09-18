package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/gontikr99/bidbot2/controller/discord"
	"github.com/gontikr99/bidbot2/controller/everquest"
	"github.com/gontikr99/bidbot2/controller/plugin"
	"log"
	"math"
	"strings"
)

func inicap(value string) string {
	if value == "" {
		return ""
	}
	value = strings.ToUpper(value[0:1]) + strings.ToLower(value[1:])
	return value
}

func RegisterDKPCommands(dc *discord.Client, eq *everquest.Client, gp *plugin.GuildPlugin) {
	eq.RegisterTellCommand("!dkp", func(who string, args string) {
		main, err := gp.GetMain(who)
		if err != nil {
			eq.Tell(who, "An error occurred, sorry.")
			log.Printf("Failed to lookup main of %v: %v", who, err)
			return
		} else if len(main) == 0 {
			eq.Tell(who, "I don't know who your main is")
			return
		}
		main = inicap(main)

		value, err := gp.GetDKP(main)
		if err != nil {
			eq.Tell(who, "An error occurred, sorry.")
			log.Printf("Failed to lookup DKP: %v", err)
		} else if math.IsNaN(value) {
			eq.Tell(who, "I don't know, sorry.")
		} else {
			eq.Tellf(who, "%v has %v DKP.", main, value)
		}
	})

	dc.RegisterDiscordCommand("!dkp", func(msg *discordgo.MessageCreate, args string) {
		dc.Fade(msg.Message)
		args = strings.TrimSpace(args)
		if args == "" {
			dc.ReplyError(msg, "dkp", "Whose DKP did you want?")
			return
		}

		main, err := gp.GetMain(strings.TrimSpace(args))
		if err != nil {
			dc.ReplyError(msg, "dkp", "An error occurred looking up the main of "+args+", sorry.")
			log.Printf("Failed to lookup main of %v: %v", args, err)
			return
		} else if len(main) == 0 {
			dc.ReplyError(msg, "dkp", "I don't know who is the main of "+args+", sorry.")
			return
		}
		main = inicap(main)

		value, err := gp.GetDKP(main)
		if err != nil {
			dc.ReplyError(msg, "dkp", "An error occurred getting the DKP of "+main+", sorry.")
			log.Printf("Failed to lookup DKP: %v", err)
		} else if math.IsNaN(value) {
			dc.ReplyWarn(msg, "dkp", "I don't know what "+main+"'s DKP total is.")
		} else {
			dc.ReplyOK(msg, "dkp", fmt.Sprintf("%v has %v DKP.", main, value))
		}
	})
}
