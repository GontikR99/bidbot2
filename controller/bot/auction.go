package bot

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/gontikr99/bidbot2/controller/assets"
	"github.com/gontikr99/bidbot2/controller/discord"
	"github.com/gontikr99/bidbot2/controller/everquest"
	"github.com/gontikr99/bidbot2/controller/plugin"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var tellRE = regexp.MustCompile("^([A-Za-z]+) (?:tells|told) you, '(.*)'$")
var numRE = regexp.MustCompile("^([-+]?(?:[0-9]*\\.?[0-9]+))(?:[^0-9].*)?$")

type bidEntry struct {
	Bidder   string
	BidText  string
	MsgEntry *discordgo.Message
}

type allBids struct {
	bidTexts []bidEntry
	bids     map[string]float64
}

func logOnError(args ...interface{}) {
	err, ok := args[len(args)-1].(error)
	if ok && err != nil {
		log.Println(err)
	}
}

// Say something in guild and in discord at the same time, and wait for both to finish happening
func announce(eqc *everquest.Client, dc *discord.Client, eqText []interface{}, dcText string) {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		logOnError(eqc.Announce(eqText...))
		wg.Done()
	}()
	go func() {
		logOnError(dc.Say(dcText))
		wg.Done()
	}()
	wg.Wait()
}

// Call into the plugin to get a solicitation for the item
func solicit(gp *plugin.GuildPlugin, itemName string) string {
	solText, err := gp.Solicit(itemName)
	if err != nil {
		log.Println(err)
		return "Bids on " + itemName + "."
	} else {
		return solText
	}
}

func RegisterAuctionCommand(eqc *everquest.Client, dc *discord.Client, gp *plugin.GuildPlugin) {
	auctionRunning := uint32(0)
	eqc.RegisterCCCommand("!auc", func(who string, args string) {
		if !atomic.CompareAndSwapUint32(&auctionRunning, 0, 1) {
			logOnError(eqc.Tellf(who, "Currently running an auction, please try later"))
			return
		}
		defer func() {
			atomic.StoreUint32(&auctionRunning, 0)
		}()

		// Find item link
		itemName := strings.TrimSpace(args)
		if len(itemName) == 0 {
			logOnError(eqc.Tell(who, "What did you want me to auction?"))
			return
		}
		itemEscape := strings.ReplaceAll(itemName, "`", "'")
		logOnError(eqc.Tellf(who, "Starting auction on %v", itemName))
		itemOffset := strings.Index(args, itemName)
		spaces := args[:itemOffset]
		itemLink, err := eqc.RaiseLink("!auc" + spaces + "{" + itemName + "}")
		if err != nil {
			logOnError(eqc.Tellf(who, "I couldn't find the item window.  Did you send a link?"))
			return
		}
		defer func() { logOnError(eqc.ClearWindows()) }()

		// Setup to collect bids
		logMessages, tapDone := eqc.TapLog()
		_, err = dc.Writef("---- [%v] **Bid Start**: `%v`", who, itemEscape)
		if err != nil {
			log.Println(err)
			logOnError(eqc.Tellf(who, "Failed to send initial message to discord: %v", err))
			return
		}
		raidDump, err := eqc.RaidDump()
		if err == nil {
			logOnError(dc.Upload("raiddump.txt", raidDump))
		}
		resultChan := make(chan *allBids)
		subCtx, subDone := context.WithCancel(eqc.Context)
		go func() {
			defer tapDone()
			result := &allBids{
				bidTexts: make([]bidEntry, 0),
				bids:     make(map[string]float64),
			}
			for {
				select {
				case <-subCtx.Done():
					resultChan <- result
					return
				case msg := <-logMessages:
					matchTell := tellRE.FindStringSubmatch(msg.Message)
					if matchTell == nil {
						continue
					}
					teller := strings.ToLower(matchTell[1])
					tellMsg := matchTell[2]
					if strings.HasPrefix(tellMsg, "!") || strings.Contains(tellMsg, "A.F.K.") || strings.Contains(tellMsg, "AFK Message") {
						continue
					}
					dmsg, err := dc.Writef("`%v` sent me a tell", inicap(teller))
					if err != nil {
						log.Println(err)
					} else {
						result.bidTexts = append(result.bidTexts, bidEntry{
							Bidder:   teller,
							BidText:  tellMsg,
							MsgEntry: dmsg,
						})
					}
					numRE := numRE.FindStringSubmatch(tellMsg)
					if numRE == nil {
						go func() {
							logOnError(eqc.Tellf(teller, "You told me \"%v\", and I can't make any sense of that as a bid.", tellMsg))
							logOnError(eqc.Tellf(teller, "Please send me your bid as a number, or 0 to cancel a previous bid."))
						}()
						continue
					}
					bidValue, err := strconv.ParseFloat(numRE[1], 64)
					if err != nil {
						log.Println(err)
						go func() { logOnError(eqc.Tellf(teller, "Sorry, I had a problem understanding your bid.")) }()
						continue
					}
					if bidValue == 0 {
						if _, ok := result.bids[teller]; ok {
							delete(result.bids, teller)
							go func() { logOnError(eqc.Tellf(teller, "Cancelled your bid.")) }()
						} else {
							go func() { logOnError(eqc.Tellf(teller, "You haven't placed a bid yet!")) }()
						}
						continue
					}
					errmsg, err := gp.ValidateBid(teller, bidValue)
					if err != nil {
						log.Println(err)
						go func() { logOnError(eqc.Tellf(teller, "Sorry, I had a problem validating your bid.")) }()
						continue
					}
					if errmsg != "" {
						go func() { logOnError(eqc.Tell(teller, errmsg)) }()
						continue
					}
					prevBid, hadPrev := result.bids[teller]
					result.bids[teller] = bidValue
					dkpTotal, err := gp.GetDKP(teller)
					if err == nil && bidValue > dkpTotal {
						go func() {
							logOnError(eqc.Tellf(teller, "Entering your bid of %v on %v, even though you only have %v DKP.  Send 0 to cancel.",
								bidValue, itemName, dkpTotal))
						}()
					} else {
						if err != nil {
							log.Println(err)
						}
						if hadPrev {
							go func() {
								logOnError(eqc.Tellf(teller, "Received your bid of %v on %v, replacing your previous bid of %v.  Send 0 to cancel.", bidValue, itemName, prevBid))
							}()
						} else {
							go func() {
								logOnError(eqc.Tellf(teller, "Received your bid of %v on %v.  Send 0 to cancel.", bidValue, itemName))
							}()
						}
					}
				}
			}
		}()

		// Talk while we've got the collector running in the background
		func() {
			defer subDone()
			err = dc.Play(assets.BellTone())
			if err != nil {
				log.Println(err)
			}
			announce(eqc, dc,
				[]interface{}{">> Bidding starts on ", itemLink, ", send me a number (to see your posted total send me !dkp).  60 seconds remain. <<"},
				solicit(gp, itemEscape)+".  60 seconds to go!")
			select {
			case <-eqc.Context.Done():
				return
			case <-time.After(30 * time.Second):
				break
			}

			announce(eqc, dc,
				[]interface{}{">> Bid for ", itemLink, ", send me a number (and only a number).  30 seconds remain. <<"},
				solicit(gp, itemEscape)+".  30 seconds to go!")
			select {
			case <-eqc.Context.Done():
				return
			case <-time.After(20 * time.Second):
				break
			}

			announce(eqc, dc,
				[]interface{}{">> Bid for ", itemLink, ", send me a number (and only a number).  10 seconds remain. <<"},
				solicit(gp, itemEscape)+".  Last call!")
			select {
			case <-eqc.Context.Done():
				return
			case <-time.After(10 * time.Second):
				break
			}

			logOnError(dc.Play(assets.BellTone()))
			announce(eqc, dc,
				[]interface{}{">> Bidding closed for ", itemLink, " <<"},
				"No more bids for "+itemEscape+".")
		}()
		auctionResult := <-resultChan
		price, winners, displays, err := gp.SortBids(auctionResult.bids, 1)
		if err != nil {
			log.Println(err)
		} else if len(winners) == 0 {
			logOnError(eqc.Announce(">> Preliminary winner(s) of ", itemLink, ": no bids <<"))
			go func() {
				logOnError(dc.WriteComplex(&discordgo.MessageSend{
					Embed: &discordgo.MessageEmbed{
						Title:       "Bid end",
						Description: fmt.Sprintf("%v: No bids", itemEscape),
						Color:       0x007f00,
					},
				}))
			}()
		} else {
			for idx, winner := range winners {
				winners[idx] = inicap(winner)
			}
			logOnError(eqc.Announce(">> Preliminary winner(s) of ", itemLink,
				": "+strings.Join(winners, " "),
				fmt.Sprintf(" for %v DKP. <<", price)))
			go func() {
				eb := &discordgo.MessageEmbed{
					Title:       "Bid end",
					Description: fmt.Sprintf("%v: [%v DKP] `%v`", itemEscape, price, strings.Join(winners, "`, `")),
					Fields:      make([]*discordgo.MessageEmbedField, 0),
					Color:       0x007f00,
				}
				for i := 0; i < 9 && i < len(displays); i++ {
					eb.Fields = append(eb.Fields, &discordgo.MessageEmbedField{
						Name:   "`" + displays[i].BidderDesc + "`",
						Value:  displays[i].BidDesc,
						Inline: true,
					})
				}
				logOnError(dc.WriteComplex(&discordgo.MessageSend{Embed: eb}))
				for i := 9; i < len(displays); i += 9 {
					eb = &discordgo.MessageEmbed{
						Title:  "Bid end",
						Fields: make([]*discordgo.MessageEmbedField, 0),
						Color:  0x007f00,
					}
					for j := i; j < i+9 && j < len(displays); j++ {
						eb.Fields = append(eb.Fields, &discordgo.MessageEmbedField{
							Name:   "`" + displays[j].BidderDesc + "`",
							Value:  displays[j].BidDesc,
							Inline: true,
						})
					}
					logOnError(dc.WriteComplex(&discordgo.MessageSend{Embed: eb}))
				}
			}()
		}
		for _, toUpdate := range auctionResult.bidTexts {
			go func(updateEntry bidEntry) {
				logOnError(dc.Session.ChannelMessageEdit(
					updateEntry.MsgEntry.ChannelID,
					updateEntry.MsgEntry.ID,
					fmt.Sprintf("`%v:` %v", inicap(updateEntry.Bidder), updateEntry.BidText)))
			}(toUpdate)
		}
	})
}
