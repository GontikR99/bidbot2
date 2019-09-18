package storage

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type cacheEntry struct {
	Timestamp time.Time
	Text      string
}

type WebCache interface {
	FetchHTTP(string, time.Duration) (string, error)
}

type DatabaseWebCache struct {
}

func (*DatabaseWebCache) FetchHTTP(url string, timeout time.Duration) (text string, err error) {
	ce := &cacheEntry{}
	err = database.Get(url, ce)
	if err == nil && ce.Timestamp.Add(timeout).After(time.Now()) {
		return ce.Text, nil
	}
	log.Printf("Fetching %v for cache", url)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	textBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil
	}
	text = string(textBytes)
	ce = &cacheEntry{
		Timestamp: time.Now(),
		Text:      text,
	}
	database.Upsert(url, ce)
	return
}
