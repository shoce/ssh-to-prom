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

	logReSubmatchCount = 4
	logReSuffix        = `(?:Invalid user|Failed password for) (\S+) from (\S+) port (\S+)`

	TsFormats = []struct {
		// https://pkg.go.dev/regexp/syntax
		timeRe *regexp.Regexp
		// https://pkg.go.dev/time
		timeFmt string
	}{
		{regexp.MustCompile(`^(\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d)\..*: ` + logReSuffix), "2006-01-02T15:04:05"},
		{regexp.MustCompile(`^(\w\w\w +\d\d? \d\d:\d\d:\d\d) .*: ` + logReSuffix), "Jan _2 15:04:05"},
	}
)

func (p failedConnEventParser) Parse(s string) (FailedConnEvent, error) {
	var timeReSm []string
	var timeFmt string

	for _, tf := range TsFormats {
		rs := tf.timeRe.FindStringSubmatch(s)
		if len(rs) == logReSubmatchCount+1 {
			timeReSm = rs
			timeFmt = tf.timeFmt
			break
		}
	}

	if timeReSm == nil {
		return FailedConnEvent{}, errWrongFormat
	}

	ts, err := time.Parse(timeFmt, timeReSm[1])
	if err != nil {
		return FailedConnEvent{}, errWrongFormat
	}

	username := timeReSm[2]
	ipaddr := net.ParseIP(timeReSm[3])

	tcpport, err := strconv.Atoi(timeReSm[4])
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
