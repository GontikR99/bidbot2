package bot

import (
	"github.com/gontikr99/bidbot2/controller/everquest"
	"log"
	"strings"
)

// Since `Client.FindTextOnScreen` isn't perfectly reliable, we have the owner get a pixel-perfect capture of some
// on-screen text to use as an anchor.
func RegisterLinkCommands(eqc *everquest.Client) {
	eqc.RegisterCCCommand("!calibrate", func(who, args string) {
		log.Println("Running C&C calibration")
		err := eqc.Calibrate()
		if err != nil {
			log.Println(err)
			eqc.Tellf(who, "A problem occurred: %v", err)
		} else {
			eqc.Tell(who, "Calibration completed.")
		}
	})

	eqc.RegisterCCCommand("!echo", func(who string, args string) {
		itemName := strings.TrimSpace(args)
		if len(itemName) == 0 {
			eqc.Tell(who, "What link did you want me to echo?")
		}
		itemOffset := strings.Index(args, itemName)
		spaces := args[:itemOffset]
		itemLink, err := eqc.RaiseLink("!echo " + spaces + "{" + itemName + "}")
		if err != nil {
			eqc.Tellf(who, "I couldn't find the item window.  Did you send a link?")
			return
		}
		eqc.Send("/tell ", who, " I see ", itemLink, ".  Do you see it?")
		return
		eqc.Tellf(who, "I couldn't find the item window.  Did you send a link?")
	})
}
