package storage

import (
	"github.com/timshannon/bolthold"
	"os"
)

var database *bolthold.Store

func init() {
	appdata := os.Getenv("APPDATA")
	bbDataDir := appdata + "\\BidBot"
	_, err := os.Stat(bbDataDir)
	if os.IsNotExist(err) {
		err := os.Mkdir(bbDataDir, 0777)
		if err != nil {
			panic(err)
		}
	}
	database, err = bolthold.Open(bbDataDir+"\\storage", 0666, nil)
	if err != nil {
		panic(err)
	}
}
