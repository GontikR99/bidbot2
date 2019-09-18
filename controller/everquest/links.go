package everquest

import (
	"errors"
	"fmt"
	"github.com/gontikr99/bidbot2/controller/assets"
	"github.com/gontikr99/bidbot2/controller/imagemanip"
	"github.com/gontikr99/bidbot2/controller/storage"
	"image"
	"sort"
	"strings"
	"time"
)

const (
	itemBoxPollDelay      = 30 * time.Millisecond
	itemBoxAttemptTimeout = 1200 * time.Millisecond
	openLinkRetryCount    = 3

	linkCopyDelay  = 50 * time.Millisecond
	linkIconDeltaX = 30
	linkIconDeltaY = 50
)

func (eqc *Client) Calibrate() error {
	if !eqc.Config.UseLinks() {
		return errors.New("No need to calibrate, bot isn't linking items.")
	}
	matches, err := eqc.FindTextOnScreen(eqc.calibrationText())
	if err != nil {
		return fmt.Errorf("Failed to perform calibration: %v", err)
	}
	if len(matches) == 0 {
		return errors.New("No calibration matches found")
	}
	sort.Sort(byYDesc(matches))
	img, err := eqc.Capture(matches[0].Rectangle)
	if err != nil {
		return fmt.Errorf("Failed to capture calibrated portion: %v", err)
	}
	eqc.Config.SetChannelImage(img)
	return nil
}

func (eqc *Client) calibrationText() string {
	return "tells " + storage.ChannelName(eqc.Config) + ":1, '"
}

func (eqc *Client) textDimensions(message string) (r image.Rectangle, err error) {
	txtImg, err := drawText(15, message)
	if err != nil {
		return
	}
	r = txtImg.Bounds()
	return
}

type byYDesc []imagemanip.MatchLocation

func (mls byYDesc) Len() int           { return len(mls) }
func (mls byYDesc) Less(i, j int) bool { return mls[i].Min.Y > mls[j].Min.Y }
func (mls byYDesc) Swap(i, j int)      { mls[i], mls[j] = mls[j], mls[i] }

func braceText(msgText string) (result string, err error) {
	braceOpen := strings.IndexRune(msgText, '{')
	if braceOpen == -1 {
		err = errors.New("No open curly brace found")
		return
	}
	braceClose := strings.IndexRune(msgText[braceOpen:], '}')
	if braceClose == -1 {
		err = errors.New("No close curly brace found")
		return
	}
	result = msgText[braceOpen+1 : braceOpen+braceClose]
	return
}

func (eqc *Client) clickLink(msgText string) error {
	chanImg := eqc.Config.ChannelImage()
	if chanImg == nil {
		return errors.New("You must run !calibrate first")
	}

	braceOpen := strings.IndexRune(msgText, '{')
	if braceOpen == -1 {
		return errors.New("No open curly brace found")
	}
	braceClose := strings.IndexRune(msgText[braceOpen:], '}')
	if braceClose == -1 {
		return errors.New("No close curly brace found")
	}
	realTextLeft := eqc.calibrationText() + msgText[:braceOpen]
	realLinkText := msgText[braceOpen+1 : braceOpen+braceClose]
	rightDims, err := eqc.textDimensions(realLinkText)
	if err != nil {
		return err
	}
	fullDims, err := eqc.textDimensions(realTextLeft + realLinkText)
	if err != nil {
		return err
	}

	height := fullDims.Dy()
	width := rightDims.Dx()
	offset := fullDims.Dx() - rightDims.Dx()

	eqi, err := eqc.GrabInput()
	if err != nil {
		return err
	}
	defer eqi.Release()

	eqi.ClearWindows()

	matches, err := eqc.FindOnScreen(chanImg)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return errors.New("Found no text matches on screen")
	}
	sort.Sort(byYDesc(matches))

	return eqi.ClickAt(matches[0].Min.X+offset+width/2, matches[0].Min.Y+height/2)
}

// Find the most recent C&C window message, click the link in it, wait for the link window to
// come up, then return a function which can be used to click on the icon in the link window.
func (eqc *Client) RaiseLink(msgText string) (linkClick func(EqInput), err error) {
	if !eqc.Config.UseLinks() {
		var bt string
		bt, err = braceText(msgText)
		if err != nil {
			return
		}
		linkClick = func(eqi EqInput) {
			typewrite(bt)
		}
		return
	}
	for i := 0; i < openLinkRetryCount; i++ {
		err = eqc.clickLink(msgText)
		if err != nil {
			return
		}
		for startTime := time.Now(); startTime.Add(itemBoxAttemptTimeout).After(time.Now()); {
			select {
			case <-eqc.Context.Done():
				err = errors.New("Client shutdown before item box appeared")
				return
			case <-time.After(itemBoxPollDelay):
				break
			}
			var matches []imagemanip.MatchLocation
			matches, err = eqc.FindOnScreen(assets.WindowCorner())
			if err != nil {
				return
			}
			for _, match := range matches {
				if match.Min.Y >= 256 {
					linkClick = func(eqi EqInput) {
						eqi.ClickAt(match.Min.X+linkIconDeltaX, match.Min.Y+linkIconDeltaY)
						time.Sleep(linkCopyDelay)
					}
					return
				}
			}
		}
	}
	err = errors.New("Never saw link box, was there really link text there?")
	return
}
