package everquest

import (
	"errors"
	"log"
	"regexp"
	"time"
)

var (
	chanRE   = regexp.MustCompile("^Channels: .*$")
	noChanRE = regexp.MustCompile("^You are not on any channels$")
	badPWRE  = regexp.MustCompile("^Incorrect password for channel .*$")
)

func leaveChannels(eqc *Client) error {
	tap, done := eqc.TapLog()
	defer done()
	for {
		err := eqc.Send("/leave 1")
		if err != nil {
			return err
		}
		timeout := time.After(5 * time.Second)
		for readyForLeave := false; !readyForLeave; {
			select {
			case msg := <-tap:
				if noChanRE.MatchString(msg.Message) {
					return nil
				}
				if chanRE.MatchString(msg.Message) {
					readyForLeave = true
				}
			case <-timeout:
				return errors.New("Failed to leave a channel")
			}
		}
	}
}

func joinChannel(eqc *Client, chantext string) error {
	tap, done := eqc.TapLog()
	defer done()
	err := eqc.Send("/join " + chantext)
	if err != nil {
		return err
	}
	timeout := time.After(3 * time.Second)
	for {
		select {
		case msg := <-tap:
			if badPWRE.MatchString(msg.Message) {
				return errors.New("Bad password")
			}
		case <-timeout:
			return nil
		}
	}
}

func checkChannels(eqc *Client) error {
	tap, done := eqc.TapLog()
	defer done()
	err := eqc.Send("/list")
	if err != nil {
		return err
	}
	timeout := time.After(5 * time.Second)
	for {
		select {
		case msg := <-tap:
			if noChanRE.MatchString(msg.Message) {
				return errors.New("Not in a channel")
			}
			if chanRE.MatchString(msg.Message) {
				return nil
			}
		case <-timeout:
			return errors.New("Never saw channel listing")
		}
	}
}

// Prepare the client for command and control functionality
func (eqc *Client) SetupCommandAndControl() error {
	log.Println("Setting up EverQuest command & control channel.")
	eqc.ClearWindows()
	err := eqc.Send("/log on")
	if err != nil {
		log.Printf("Failed to turn logs on: %v", err)
		return err
	}
	err = leaveChannels(eqc)
	if err != nil {
		log.Printf("Failed to leave channels: %v", err)
		return err
	}
	err = joinChannel(eqc, eqc.Config.ChatChannelAndPassword())
	if err != nil {
		log.Printf("Failed to join C&C channel: %v", err)
		return err
	}
	err = checkChannels(eqc)
	if err != nil {
		log.Printf("No channels present", err)
		return err
	}
	return nil
}
