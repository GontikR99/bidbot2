package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/gontikr99/bidbot2/controller/discord"
	"github.com/gontikr99/bidbot2/controller/storage"
)

func RegisterDiscordBindCommands(dc *discord.Client) {
	dc.RegisterDiscordCommand("!bindvoice", func(m *discordgo.MessageCreate, args string) {
		dc.Fade(m.Message)
		if !dc.IsFromAdmin(m) {
			dc.ReplyError(m, "bindvoice", "Only server admins may !bindvoice")
			return
		}

		// Find the channel that the message came from.
		c, err := dc.Session.State.Channel(m.ChannelID)
		if err != nil {
			dc.ReplyError(m, "bindvoice", "Failed to retrieve channel")
			return
		}

		// Find the guild for that channel.
		g, err := dc.Session.State.Guild(c.GuildID)
		if err != nil {
			dc.ReplyError(m, "bindvoice", "Failed to find guild")
			return
		}

		// Look for the message sender in that guild's current voice states.
		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {
				dc.Config.SetVoiceChannel(&storage.VoiceChannel{
					GuildID:   g.ID,
					ChannelID: vs.ChannelID,
				})
				dc.ReplyOK(m, "bindvoice", "Voice channel bound")
				return
			}
		}
		dc.ReplyError(m, "bindvoice", "Failed to set voice channel: Are you in a channel?")
	})

	dc.RegisterDiscordCommand("!bindtext", func(msg *discordgo.MessageCreate, args string) {
		dc.Fade(msg.Message)
		if !dc.IsFromAdmin(msg) {
			dc.ReplyError(msg, "bindtext", "Only server admins may !bind")
			return
		}

		dc.Config.SetTextChannel(msg.ChannelID)
		dc.ReplyOK(msg, "bindtext", "Channel bound")
	})

	dc.RegisterDiscordCommand("!unbind", func(msg *discordgo.MessageCreate, args string) {
		dc.Fade(msg.Message)
		if !dc.IsFromAdmin(msg) {
			dc.ReplyError(msg, "unbind", "Only server admins may !bind")
			return
		}

		dc.Config.SetTextChannel("")
		dc.ReplyOK(msg, "bind", "Channel unbound")
	})
}
