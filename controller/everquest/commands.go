package everquest

import (
	"github.com/gontikr99/bidbot2/controller/storage"
	"regexp"
)

// Start a command handler which watches the C&C channel
func (eqc *Client) RegisterCCCommand(command string, callback func(string, string)) {
	rxp := regexp.MustCompile("^([A-Za-z]+) tells " + storage.ChannelName(eqc.Config) + ":1, '" + command + "((?:\\s+.*))?'$")
	go func() {
		tap, done := eqc.TapLog()
		defer done()
		for {
			select {
			case msg := <-tap:
				parts := rxp.FindStringSubmatch(msg.Message)
				if parts != nil {
					who := parts[1]
					args := parts[2]
					go callback(who, args)
				}
			case <-eqc.Context.Done():
				return
			}
		}
	}()
}

// Start a command handler which watches tells
func (eqc *Client) RegisterTellCommand(command string, callback func(string, string)) {
	rxp := regexp.MustCompile("^([A-Za-z]+) (?:tells|told) you, '" + command + "((?:\\s+.*))?'$")
	go func() {
		tap, done := eqc.TapLog()
		defer done()
		for {
			select {
			case msg := <-tap:
				parts := rxp.FindStringSubmatch(msg.Message)
				if parts != nil {
					who := parts[1]
					args := parts[2]
					go callback(who, args)
				}
			case <-eqc.Context.Done():
				return
			}
		}
	}()
}
