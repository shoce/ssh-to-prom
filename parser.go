package main

import (
	"errors"
	"net"
	"regexp"
	"strconv"
	"time"
)

type EventParser interface {
	Parse(s string) (*ConnEvent, error)
}

type ConnEventParser struct{}

var (
	errWrongFormat = errors.New("wrong event format")

	logMsgRe     = `(Accepted|Failed) (password|publickey) for (?:invalid user |)(\S+) from (\S+) port (\S+)`
	logMsgRegexp = regexp.MustCompile(logMsgRe)

	logFormats = []struct {
		// https://pkg.go.dev/regexp/syntax
		logRe *regexp.Regexp
		// https://pkg.go.dev/time
		timeFmt string
	}{
		{regexp.MustCompile(`^(\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d)\.\S+ \S+ \S+: ` + logMsgRe), "2006-01-02T15:04:05"},
		{regexp.MustCompile(`^(\w\w\w +\d\d? \d\d:\d\d:\d\d) \S+ \S+: ` + logMsgRe), "Jan _2 15:04:05"},
	}
)

func (p ConnEventParser) Parse(s string) (*ConnEvent, error) {
	if !logMsgRegexp.MatchString(s) {
		return nil, nil
	}

	var logReSm []string
	var timeFmt string

	for _, logfmt := range logFormats {
		logresm := logfmt.logRe.FindStringSubmatch(s)
		if len(logresm) == logfmt.logRe.NumSubexp()+1 {
			logReSm = logresm
			timeFmt = logfmt.timeFmt
			break
		}
	}

	if logReSm == nil {
		return &ConnEvent{}, errWrongFormat
	}

	ts, err := time.Parse(timeFmt, logReSm[1])
	if err != nil {
		return &ConnEvent{}, errWrongFormat
	}

	accepted := logReSm[2] == "Accepted"
	authmethod := logReSm[3]
	user := logReSm[4]
	addr := net.ParseIP(logReSm[5])

	port, err := strconv.Atoi(logReSm[6])
	if err != nil {
		return &ConnEvent{}, errWrongFormat
	}

	// The logs do not have information about the year, so we're just assuming we're parsing current year logs
	ts = time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second(), ts.Nanosecond(), time.UTC)

	return &ConnEvent{
		Timestamp:  ts,
		Accepted:   accepted,
		AuthMethod: authmethod,
		User:       user,
		Addr:       addr,
		Port:       port,
	}, nil
}
