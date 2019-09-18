package storage

import (
	"bytes"
	"image"
	"image/png"
	"log"
	"strconv"
)

type BoltholdBackedConfig struct {
}

type bhConfigEntry struct {
	Data []byte
}

const (
	eqDirKey        = "eqDir"
	charNameKey     = "charName"
	chatChanKey     = "chatChan"
	chanImgKey      = "chanImg"
	discordTokenKey = "discordToken"
	cloudTTSCredKey = "cloutTTSCred"
	textChannelKey  = "textChannel"
	voiceChannelKey = "voiceChannel"
	rulesLuaKey     = "rulesLua"
	useLinksKey     = "useLinks"
	announceChanKey = "announceChannel"
)

func (bhc *BoltholdBackedConfig) VoiceChannel() *VoiceChannel {
	value := &VoiceChannel{}
	err := database.Get(voiceChannelKey, value)
	if err != nil {
		return nil
	} else {
		return value
	}
}

func (bhc *BoltholdBackedConfig) SetVoiceChannel(vc *VoiceChannel) {
	database.Upsert(voiceChannelKey, vc)
}

func (bhc *BoltholdBackedConfig) EverQuestDirectory() string {
	value := &bhConfigEntry{}
	err := database.Get(eqDirKey, value)
	if err == nil {
		return string(value.Data)
	} else {
		return "C:\\Users\\Public\\Daybreak Game Company\\Installed Games\\EverQuest"
	}
}

func (bhc *BoltholdBackedConfig) SelectedCharacter() string {
	value := &bhConfigEntry{}
	err := database.Get(charNameKey, value)
	if err == nil {
		return string(value.Data)
	} else {
		return ""
	}
}

func (bhc *BoltholdBackedConfig) ChatChannelAndPassword() string {
	value := &bhConfigEntry{}
	err := database.Get(chatChanKey, value)
	if err == nil {
		return string(value.Data)
	} else {
		return ""
	}
}

func (bhc *BoltholdBackedConfig) ChannelImage() image.Image {
	value := &bhConfigEntry{}
	err := database.Get(chanImgKey, value)
	if err != nil {
		return nil
	}
	buf := bytes.NewBuffer(value.Data)
	img, err := png.Decode(buf)
	if err != nil {
		return nil
	}
	return img
}

func (bhc *BoltholdBackedConfig) SetEverQuestDirectory(eqDir string) {
	err := database.Upsert(eqDirKey, &bhConfigEntry{[]byte(eqDir)})
	if err != nil {
		log.Println(err)
	}
}

func (bhc *BoltholdBackedConfig) SetSelectedCharacter(charName string) {
	err := database.Upsert(charNameKey, &bhConfigEntry{[]byte(charName)})
	if err != nil {
		log.Println(err)
	}
}

func (bhc *BoltholdBackedConfig) SetChatChannelAndPassword(chatChan string) {
	err := database.Upsert(chatChanKey, &bhConfigEntry{[]byte(chatChan)})
	if err != nil {
		log.Println(err)
	}
}

func (bhc *BoltholdBackedConfig) SetChannelImage(chanImg image.Image) {
	if chanImg == nil {
		database.Delete(chanImgKey, &bhConfigEntry{})
		return
	}
	buffer := &bytes.Buffer{}
	err := png.Encode(buffer, chanImg)
	if err != nil {
		log.Println(err)
	}
	err = database.Upsert(chanImgKey, &bhConfigEntry{buffer.Bytes()})
	if err != nil {
		log.Println(err)
	}
}

func (bhc *BoltholdBackedConfig) DiscordToken() string {
	value := &bhConfigEntry{}
	err := database.Get(discordTokenKey, value)
	if err == nil {
		return string(value.Data)
	} else {
		return ""
	}
}

func (bhc *BoltholdBackedConfig) CloudTTSCredPath() string {
	value := &bhConfigEntry{}
	err := database.Get(cloudTTSCredKey, value)
	if err == nil {
		return string(value.Data)
	} else {
		return ""
	}
}

func (bhc *BoltholdBackedConfig) SetDiscordToken(value string) {
	err := database.Upsert(discordTokenKey, &bhConfigEntry{[]byte(value)})
	if err != nil {
		log.Println(err)
	}
}

func (bhc *BoltholdBackedConfig) SetCloudTTSCredPath(value string) {
	err := database.Upsert(cloudTTSCredKey, &bhConfigEntry{[]byte(value)})
	if err != nil {
		log.Println(err)
	}
}

func (bhc *BoltholdBackedConfig) RulesLua() string {
	value := &bhConfigEntry{}
	err := database.Get(rulesLuaKey, value)
	if err == nil {
		return string(value.Data)
	} else {
		return ""
	}
}

func (bhc *BoltholdBackedConfig) SetRulesLua(value string) {
	err := database.Upsert(rulesLuaKey, &bhConfigEntry{[]byte(value)})
	if err != nil {
		log.Println(err)
	}
}

func (bhc *BoltholdBackedConfig) TextChannel() string {
	value := &bhConfigEntry{}
	err := database.Get(textChannelKey, value)
	if err == nil {
		return string(value.Data)
	} else {
		return ""
	}
}

func (bhc *BoltholdBackedConfig) SetTextChannel(value string) {
	var err error
	if value == "" {
		err = database.Delete(textChannelKey, &bhConfigEntry{})
	} else {
		err = database.Upsert(textChannelKey, &bhConfigEntry{[]byte(value)})
	}
	if err != nil {
		log.Println(err)
	}
}

func (bhc *BoltholdBackedConfig) UseLinks() bool {
	value := &bhConfigEntry{}
	err := database.Get(useLinksKey, value)
	if err == nil {
		result, err2 := strconv.ParseBool(string(value.Data))
		return err2 == nil && result
	} else {
		return false
	}
}

func (bhc *BoltholdBackedConfig) SetUseLinks(value bool) {
	err := database.Upsert(useLinksKey, &bhConfigEntry{[]byte(strconv.FormatBool(value))})
	if err != nil {
		log.Println(err)
	}
}

func (bhc *BoltholdBackedConfig) AnnounceChannel() string {
	value := &bhConfigEntry{}
	err := database.Get(announceChanKey, value)
	if err == nil {
		return string(value.Data)
	} else {
		return "/gu"
	}
}

func (bhc *BoltholdBackedConfig) SetAnnounceChannel(value string) {
	err := database.Upsert(announceChanKey, &bhConfigEntry{[]byte(value)})
	if err != nil {
		log.Println(err)
	}
}
