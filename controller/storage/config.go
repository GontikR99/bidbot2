package storage

import (
	"image"
	"strings"
)

type ControllerConfig interface {
	// Settings established before startup
	AnnounceChannel() string
	EverQuestDirectory() string
	SelectedCharacter() string
	ChatChannelAndPassword() string
	DiscordToken() string
	CloudTTSCredPath() string
	RulesLua() string
	UseLinks() bool

	SetAnnounceChannel(string)
	SetEverQuestDirectory(string)
	SetSelectedCharacter(string)
	SetChatChannelAndPassword(string)
	SetDiscordToken(string)
	SetCloudTTSCredPath(string)
	SetRulesLua(string)
	SetUseLinks(bool)

	// Settings established during run
	ChannelImage() image.Image
	TextChannel() string
	VoiceChannel() *VoiceChannel

	SetChannelImage(image.Image)
	SetTextChannel(string)
	SetVoiceChannel(*VoiceChannel)
}

type VoiceChannel struct {
	GuildID   string
	ChannelID string
}

func ChannelName(cc ControllerConfig) string {
	cpws := cc.ChatChannelAndPassword()
	idx := strings.IndexByte(cpws, ':')
	if idx > 0 {
		return cpws[:idx]
	} else {
		return ""
	}
}
