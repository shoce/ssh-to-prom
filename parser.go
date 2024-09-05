package main

import (
	"errors"
	"net"
	"regexp"
	"strconv"
	"time"
)

// EventParser provides the logic to map from a raw event to a FailedConnEvent
type EventParser interface {
	Parse(s string) (FailedConnEvent, error)
}

// NewFailedConnEventParser returns an implementation of EventParser
func NewFailedConnEventParser() EventParser {
	return failedConnEventParser{}
}

type failedConnEventParser struct{}

var (
	errWrongFormat = errors.New("wrong event format")
	eRegex         = regexp.MustCompile(`^(?<ts>[-0-9T]+).*: Invalid user (?<user>\w+) from (?<addr>[^ ]+) port (?<port>[^ ]+)`)
)

func (p failedConnEventParser) Parse(s string) (FailedConnEvent, error) {
	rs := eRegex.FindStringSubmatch(s)
	if len(rs) != 6 {
		return FailedConnEvent{}, errWrongFormat
	}

	portNumber, err := strconv.Atoi(rs[4])
	if err != nil {
		return FailedConnEvent{}, errWrongFormat
	}

	ts, err := time.Parse("2006-01-02T15:04:05", rs[1])
	if err != nil {
		return FailedConnEvent{}, errWrongFormat
	}

	// The logs do not have information about the year, so we're just assuming we're parsing current year logs
	ts = time.Date(time.Now().Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second(), ts.Nanosecond(), time.UTC)

	return FailedConnEvent{
		Username:  rs[2],
		IPAddress: net.ParseIP(rs[3]),
		Port:      portNumber,
		Timestamp: ts,
		Country:   "unknown",
	}, nil
}
