package bot

import (
	"github.com/gontikr99/bidbot2/controller/assets"
	"github.com/gontikr99/bidbot2/controller/discord"
	"github.com/gontikr99/bidbot2/controller/everquest"
)

func RegisterSayCommands(eqc *everquest.Client, dc *discord.Client) {
	eqc.RegisterCCCommand("!say", func(who, what string) {
		dc.Say(what)
	})
	eqc.RegisterCCCommand("!announce", func(who, what string) {
		dc.Play(assets.BellTone())
		dc.Say(what)
	})
}
