package everquest

// logread.go: Watch for old and new EverQuest log files, read and parse them.

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"
)

type EqLogEntry struct {
	Character string
	Server    string
	Timestamp string
	Message   string
}

var (
	filenameMatch = regexp.MustCompile("^eqlog_([A-Za-z]*)_([A-Za-z]*).txt$")
	loglineMatch  = regexp.MustCompile("^\\[([^\\]]*)] (.*)$")
)

func readAllLogs(ctx context.Context, directory string) (<-chan EqLogEntry, error) {
	_, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	receiver := make(chan EqLogEntry, 16)
	go readAllLogsLoop(ctx, directory, receiver)
	return receiver, nil
}

func readAllLogsLoop(ctx context.Context, directory string, receiver chan<- EqLogEntry) {
	seen := make(map[string]bool)
	for {
		entries, err := ioutil.ReadDir(directory)
		if err == nil {
			for _, fi := range entries {
				if _, ok := seen[fi.Name()]; !ok {
					seen[fi.Name()] = true
					if parts := filenameMatch.FindStringSubmatch(fi.Name()); parts != nil {
						go tailLog(ctx, directory+"/"+fi.Name(), parts[1], parts[2], receiver)
					}
				}
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(200 * time.Millisecond):
		}
	}
}

func tailLog(ctx context.Context, filename string, character string, server string, receiver chan<- EqLogEntry) {
	fd, err := os.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	fd.Seek(-1, io.SeekEnd)
	rdbuf := make([]byte, 1024)
	buffer := make([]byte, 0)
	for {
		cnt, _ := fd.Read(rdbuf)
		if cnt <= 0 {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(5 * time.Millisecond)
				continue
			}
		}
		buffer = append(buffer, rdbuf[0:cnt]...)
		ib := bytes.IndexByte(buffer, '\n')
		if ib >= 0 {
			buffer = append(make([]byte, 0), buffer[ib+1:]...)
			break
		}
	}

	for {
		for ib := bytes.IndexByte(buffer, '\n'); ib >= 0; ib = bytes.IndexByte(buffer, '\n') {
			line := strings.ReplaceAll(string(buffer[0:ib-1]), "\r", "")
			buffer = append(make([]byte, 0), buffer[ib+1:]...)
			if parts := loglineMatch.FindStringSubmatch(line); parts != nil {
				receiver <- EqLogEntry{
					Character: character,
					Server:    server,
					Timestamp: parts[1],
					Message:   parts[2],
				}
			}
		}
		cnt, _ := fd.Read(rdbuf)
		if cnt <= 0 {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(5 * time.Millisecond)
				continue
			}
		}
		buffer = append(buffer, rdbuf[0:cnt]...)
	}
}
