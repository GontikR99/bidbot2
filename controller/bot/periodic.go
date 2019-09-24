package bot

import (
	"github.com/gontikr99/bidbot2/controller/discord"
	"github.com/gontikr99/bidbot2/controller/everquest"
	"log"
	"time"
)

func StartPeriodicRaidDumps(eqc *everquest.Client, dc *discord.Client) {
	go func() {
		lastDump := time.Time{}
		for {
			select {
				case <-eqc.Context.Done():
					return
				case <-time.After(10*time.Second):
			}
			nowTime:=time.Now()
			if nowTime.Minute() % 30 !=0 {
				continue
			}
			if nowTime.Hour()==lastDump.Hour() && nowTime.Minute()==lastDump.Minute() {
				continue
			}
			raidDump, err := eqc.RaidDump()
			if err!=nil {
				log.Println(err)
				continue
			}
			lastDump = nowTime
			if len(raidDump)==0 {
				log.Println("Empty raid dump, skipping")
				continue
			}
			logOnError(dc.Writef("[%d:%02d] Current raid members:", nowTime.Hour(), nowTime.Minute()))
			logOnError(dc.Upload("raiddump.txt", raidDump))
		}
	}()
}
