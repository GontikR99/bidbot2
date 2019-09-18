package everquest

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type GuildRecordsReader interface {
	GuildRecords() (gr map[string]*GuildRecord, err error)
}

const (
	dumpReadTimeout = 2 * time.Second
	dumpRetryCount  = 3
	windowOpenTime  = 400 * time.Millisecond
)

func (eqi *EqInput) blinkGuildWindow() {
	pressKey(vkMenu)
	time.Sleep(tapDelay)
	tap('g')
	releaseKey(vkMenu)
	time.Sleep(tapDelay + windowOpenTime)

	pressKey(vkMenu)
	time.Sleep(tapDelay)
	tap('g')
	releaseKey(vkMenu)
	time.Sleep(tapDelay)
}

func (eqi *EqInput) blinkRaidWindow() {
	pressKey(vkMenu)
	time.Sleep(tapDelay)
	tap('r')
	releaseKey(vkMenu)
	time.Sleep(tapDelay + windowOpenTime)

	pressKey(vkMenu)
	time.Sleep(tapDelay)
	tap('r')
	releaseKey(vkMenu)
	time.Sleep(tapDelay)
}

var outputfileComplete = regexp.MustCompile("^Outputfile Complete: (.+)$")

func readGuildRecords(filename string) (records map[string]*GuildRecord, err error) {
	fileText, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	records = make(map[string]*GuildRecord, 0)
	for _, line := range strings.Split(string(fileText), "\r\n") {
		fields := strings.Split(line, "\t")
		if len(fields) < 8 {
			continue
		}
		lvl, _ := strconv.Atoi(fields[1])
		records[strings.ToLower(fields[0])] = &GuildRecord{
			Level:      lvl,
			Class:      strings.ToLower(fields[2]),
			Rank:       strings.ToLower(fields[3]),
			IsAlt:      fields[4] == "A",
			LastOnline: fields[5],
			GuildNote:  fields[7],
		}
	}
	os.Remove(filename)
	log.Printf("Parsed guild dump, found %v members", len(records))
	if len(records) == 0 {
		err = errors.New("Guild seems kinda small to me.")
	}
	return
}

func (eqc *Client) GuildRecords() (gr map[string]*GuildRecord, err error) {
	var errt error
	for i := 0; i < dumpRetryCount; i++ {
		grt, errt := eqc.guildRecordsAttempt()
		if errt == nil {
			gr = grt
			return
		}
	}
	err = errt
	return
}

func (eqc *Client) guildRecordsAttempt() (gr map[string]*GuildRecord, err error) {
	eqc.guildRecordsSync.Lock()

	if eqc.guildRecords != nil && eqc.guildRecordTimestamp.Add(5*time.Minute).After(time.Now()) {
		defer eqc.guildRecordsSync.Unlock()
		return eqc.guildRecords, nil
	}
	eqc.guildRecordsSync.Unlock()

	eqi, err := eqc.GrabInput()
	if err != nil {
		return
	}
	tap, tapDone := eqc.TapLog()
	defer tapDone()
	eqi.blinkGuildWindow()
	eqi.Send("/outputfile guild")
	for startTime := time.Now(); startTime.Add(dumpReadTimeout).After(time.Now()); {
		select {
		case <-eqc.Context.Done():
			err = errors.New("Shutting down")
			eqi.Release()
			return
		case msg := <-tap:
			parts := outputfileComplete.FindStringSubmatch(msg.Message)
			if len(parts) == 0 {
				break
			}
			eqi.Release()
			filename := eqc.Config.EverQuestDirectory() + "\\" + parts[1]
			var records map[string]*GuildRecord
			time.Sleep(250 * time.Millisecond) // let EQ close the file.
			records, err = readGuildRecords(filename)
			if err != nil {
				return
			}

			gr = records

			eqc.guildRecordsSync.Lock()
			defer eqc.guildRecordsSync.Unlock()
			eqc.guildRecords = records
			eqc.guildRecordTimestamp = time.Now()
			return
		case <-time.After(10 * time.Millisecond):
			break
		}
	}
	err = errors.New("Never saw outputfile message")
	eqi.Release()
	return
}

type GuildRecord struct {
	Present    bool
	Level      int
	Class      string
	Rank       string
	IsAlt      bool
	LastOnline string
	GuildNote  string
}

func (eqc *Client) RaidDump() (raidData []byte, err error) {
	eqi, err := eqc.GrabInput()
	if err != nil {
		return
	}
	tap, tapDone := eqc.TapLog()
	defer tapDone()
	eqi.blinkRaidWindow()
	eqi.Send("/outputfile raid")
	for startTime := time.Now(); startTime.Add(dumpReadTimeout).After(time.Now()); {
		select {
		case <-eqc.Context.Done():
			err = errors.New("Shutting down")
			eqi.Release()
			return
		case msg := <-tap:
			parts := outputfileComplete.FindStringSubmatch(msg.Message)
			if len(parts) == 0 {
				break
			}
			eqi.Release()
			filename := eqc.Config.EverQuestDirectory() + "\\" + parts[1]
			time.Sleep(250 * time.Millisecond) // let EQ close the file.
			raidData, err = ioutil.ReadFile(filename)
			if err != nil {
				os.Remove(filename)
				return
			}

			return
		case <-time.After(10 * time.Millisecond):
			break
		}
	}
	err = errors.New("Never saw outputfile message")
	eqi.Release()
	return
}