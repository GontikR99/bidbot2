package discord

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/gontikr99/bidbot2/controller/soundmanip"
	"github.com/gontikr99/bidbot2/controller/storage"
	"log"
	"strings"
	"sync"
	"time"
)

type Client struct {
	Context context.Context
	Config  storage.ControllerConfig
	Session *discordgo.Session

	chatChan    <-chan *discordgo.MessageCreate
	chatSync    sync.Mutex
	nextChatTap int
	chatTaps    map[int]*tapPair

	voice chan<- *voiceRequest

	cleanup sync.WaitGroup
}

func NewDiscordClient(ctx context.Context, config storage.ControllerConfig) (client *Client, err error) {
	result := &Client{}
	result.Context = ctx
	result.Config = config
	result.Session, err = discordgo.New("Bot " + config.DiscordToken())
	if err != nil {
		return
	}

	chatChan := make(chan *discordgo.MessageCreate)
	result.chatChan = chatChan
	result.chatTaps = make(map[int]*tapPair)
	result.Session.AddHandler(func(s *discordgo.Session, mc *discordgo.MessageCreate) {
		chatChan <- mc
	})

	log.Println("Connecting to Discord")
	err = result.Session.Open()
	if err != nil {
		return
	}
	log.Printf("Connected to Discord.  Use this link to invite: https://discordapp.com/oauth2/authorize?client_id=%v&scope=bot&permissions=8", result.Session.State.User.ID)
	log.Println("Starting Discord chat loop")
	result.cleanup.Add(1)
	go result.forwardLogMessages()

	log.Println("Starting Discord voice loop")
	voiceChan := make(chan *voiceRequest)
	result.cleanup.Add(1)
	go result.speechLoop(voiceChan)
	result.voice = voiceChan

	go func() {
		result.cleanup.Wait()
		log.Println("Shutting down Discord")
		result.Session.Close()
	}()

	client = result
	return
}

type tapPair struct {
	messages chan<- *discordgo.MessageCreate
	done     <-chan struct{}
}

// Receive messages from the log, send them on to all taps.  Maintain taps, too.
func (dclient *Client) forwardLogMessages() {
	for {
		select {
		case msg := <-dclient.chatChan:
			dclient.chatSync.Lock()
			for k, tap := range dclient.chatTaps {
				select {
				case <-tap.done:
					delete(dclient.chatTaps, k)
				default:
					select {
					case tap.messages <- msg:
					default:
						log.Printf("Tap %d is full", k)
					}
				}
			}
			dclient.chatSync.Unlock()
		case <-dclient.Context.Done():
			log.Println("Shutting down Discord message loop")
			dclient.cleanup.Done()
			return
		}
	}
}

// Allocate a new tap, receiving all new messages from the client from this point on.  When messages are
// no longer required, send something to `done`
func (dclient *Client) TapChat() (messages <-chan *discordgo.MessageCreate, done func()) {
	mc := make(chan *discordgo.MessageCreate, 4096)
	dc := make(chan struct{}, 2)
	entry := tapPair{mc, dc}
	dclient.chatSync.Lock()
	dclient.chatTaps[dclient.nextChatTap] = &entry
	dclient.nextChatTap += 1
	dclient.chatSync.Unlock()
	return mc, func() { dc <- struct{}{} }
}

func (dclient *Client) Fade(mc *discordgo.Message) {
	dclient.cleanup.Add(1)
	go func() {
		select {
		case <-dclient.Context.Done():
			break
		case <-time.After(2 * time.Minute):
			dclient.Session.ChannelMessageDelete(mc.ChannelID, mc.ID)
		}
		dclient.cleanup.Done()
	}()
}

func (dclient *Client) ReplyOK(cmd *discordgo.MessageCreate, title string, text string) error {
	msg, err := dclient.Session.ChannelMessageSendComplex(cmd.ChannelID, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title:       title,
			Description: text,
			Color:       0x007f00,
		},
	})
	dclient.Fade(msg)
	return err
}

func (dclient *Client) ReplyWarn(cmd *discordgo.MessageCreate, title string, text string) error {
	msg, err := dclient.Session.ChannelMessageSendComplex(cmd.ChannelID, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title:       title,
			Description: text,
			Color:       0x7f7f00,
		},
	})
	dclient.Fade(msg)
	return err
}

func (dclient *Client) ReplyError(cmd *discordgo.MessageCreate, title string, text string) error {
	msg, err := dclient.Session.ChannelMessageSendComplex(cmd.ChannelID, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title:       title,
			Description: text,
			Color:       0x7f0000,
		},
	})
	dclient.Fade(msg)
	return err
}

func (dclient *Client) IsFromAdmin(msg *discordgo.MessageCreate) bool {
	guild, err := dclient.Session.Guild(msg.GuildID)
	if err != nil {
		log.Println("Failed to look up guild")
		return false
	}
	if strings.Compare(msg.Author.ID, guild.OwnerID) == 0 {
		return true
	}
	for _, role := range guild.Roles {
		if role.Permissions&discordgo.PermissionAdministrator == 0 {
			continue
		}
		for _, m := range guild.Members {
			if strings.Compare(m.User.ID, msg.Author.ID) != 0 {
				continue
			}
			for _, r := range m.Roles {
				if strings.Compare(r, role.ID) == 0 {
					return true
				}
			}
		}
	}
	return false
}

// Write text to the default (bound) channel
func (dclient *Client) Write(text string) (*discordgo.Message, error) {
	chanID := dclient.Config.TextChannel()
	if chanID == "" {
		return nil, errors.New("No channel selected yet")
	}
	return dclient.Session.ChannelMessageSend(chanID, text)
}

func (dclient *Client) Upload(name string, data []byte) (*discordgo.Message, error) {
	chanID := dclient.Config.TextChannel()
	if chanID == "" {
		return nil, errors.New("No channel selected yet")
	}
	return dclient.Session.ChannelFileSend(chanID, name, bytes.NewReader(data))
}

func (dclient *Client) Writef(fmtstr string, args ...interface{}) (*discordgo.Message, error) {
	return dclient.Write(fmt.Sprintf(fmtstr, args...))
}

func (dclient *Client) WriteComplex(msg *discordgo.MessageSend) (*discordgo.Message, error) {
	chanID := dclient.Config.TextChannel()
	if chanID == "" {
		return nil, errors.New("No channel selected yet")
	}
	return dclient.Session.ChannelMessageSendComplex(chanID, msg)
}

func (dclient *Client) Say(text string) error {
	opus, err := soundmanip.Synthesize(dclient.Config.CloudTTSCredPath(), text)
	if err != nil {
		return err
	}
	return dclient.Play(opus)
}

func (dclient *Client) Play(opus *soundmanip.OpusFile) error {
	errChan := make(chan voiceResponse)
	vr := &voiceRequest{
		sound:  opus,
		result: errChan,
	}
	dclient.voice <- vr
	select {
	case <-dclient.Context.Done():
		return errors.New("Shutdown while waiting for speech")
	case err := <-errChan:
		return err.err
	}
}

type voiceRequest struct {
	sound  *soundmanip.OpusFile
	result chan<- voiceResponse
}

type voiceResponse struct {
	err error
}

func play(vc *discordgo.VoiceConnection, opus *soundmanip.OpusFile) error {
	vc.Speaking(true)
	// Send the buffer data.
	for {
		buf, err := opus.ReadPacket()
		if err != nil {
			break
		}
		vc.OpusSend <- buf
	}

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)
	return nil
}

func (dclient *Client) speechLoop(reqChan <-chan *voiceRequest) {
	for {
		var (
			voiceClient *discordgo.VoiceConnection
			err         error
		)
		select {
		case <-dclient.Context.Done():
			log.Println("Shutting down Discord voice loop")
			dclient.cleanup.Done()
			return
		case msg := <-reqChan:
			vchan := dclient.Config.VoiceChannel()
			if vchan == nil {
				msg.result <- voiceResponse{errors.New("No voice channel set yet.  Join one and say !bindvoice in Discord.")}
				break
			}
			voiceClient, err = dclient.Session.ChannelVoiceJoin(vchan.GuildID, vchan.ChannelID, false, true)
			if err != nil {
				msg.result <- voiceResponse{fmt.Errorf("Failed to join voice channel: %v", err)}
				break
			}
			// Sleep for a specificed amount of time before starting.
			time.Sleep(250 * time.Millisecond)

			err = play(voiceClient, msg.sound)
			msg.result <- voiceResponse{err}
		}

		for goToSleep := false; !goToSleep; {
			select {
			case <-dclient.Context.Done():
				log.Println("Shutting down Discord voice loop")
				voiceClient.Disconnect()
				voiceClient.Close()
				dclient.cleanup.Done()
				return
			case <-time.After(45 * time.Second):
				log.Println("Leaving voice")
				voiceClient.Disconnect()
				voiceClient.Close()
				goToSleep = true
				break
			case msg := <-reqChan:
				err = play(voiceClient, msg.sound)
				msg.result <- voiceResponse{err}
				if err != nil {
					voiceClient.Disconnect()
					voiceClient.Close()
					goToSleep = true
				}
				break
			}
		}
	}
}
