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

	eRegexpSuffix = `(?:Invalid user|Failed password for) (\S+) from (\S+) port (\S+)`

	// https://pkg.go.dev/regexp/syntax
	eRegexp1 = regexp.MustCompile(`^(\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d)\..*: ` + eRegexpSuffix)
	// https://pkg.go.dev/time
	eTsFmt1 = "2006-01-02T15:04:05"

	// https://pkg.go.dev/regexp/syntax
	eRegexp2 = regexp.MustCompile(`^(\w\w\w +\d\d? \d\d:\d\d:\d\d) .*: ` + eRegexpSuffix)
	eTsFmt2  = "Jan _2 15:04:05"
)

func (p failedConnEventParser) Parse(s string) (FailedConnEvent, error) {
	rs := eRegexp1.FindStringSubmatch(s)
	if len(rs) != 5 {
		rs = eRegexp2.FindStringSubmatch(s)
		if len(rs) != 5 {
			return FailedConnEvent{}, errWrongFormat
		}
	}

	ts, err := time.Parse(eTsFmt1, rs[1])
	if err != nil {
		ts, err = time.Parse(eTsFmt2, rs[1])
		if err != nil {
			return FailedConnEvent{}, errWrongFormat
		}
	}

	username := rs[2]
	ipaddr := net.ParseIP(rs[3])

	tcpport, err := strconv.Atoi(rs[4])
	if err != nil {
		return FailedConnEvent{}, errWrongFormat
	}

	// The logs do not have information about the year, so we're just assuming we're parsing current year logs
	ts = time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second(), ts.Nanosecond(), time.UTC)

	return FailedConnEvent{
		Username:  username,
		IPAddress: ipaddr,
		Port:      tcpport,
		Timestamp: ts,
		Country:   "unknown",
	}, nil
}
