package main

import (
	"io"

	tail "github.com/papertrail/go-tail/follower"
)

type AsyncEventReader interface {
	Start()
	Stop()
}

type ReaderOption interface {
	Apply(ConnEvent) (ConnEvent, error)
}

type fileReader struct {
	parser    EventParser
	eventChan chan ConnEvent
	errorChan chan error
	done      chan bool
	options   []ReaderOption
	filename  string
}

func NewFileReader(filename string, parser EventParser, eventChan chan ConnEvent, errorChan chan error, options ...ReaderOption) AsyncEventReader {
	done := make(chan bool)
	return fileReader{
		filename:  filename,
		parser:    parser,
		eventChan: eventChan,
		errorChan: errorChan,
		done:      done,
		options:   options,
	}
}

func (fr fileReader) Stop() {
	fr.done <- true
}

func (fr fileReader) Start() {
	t, err := tail.New(fr.filename, tail.Config{
		Whence: io.SeekStart,
		Offset: 0,
		Reopen: true,
	})
	if err != nil {
		panic("ERROR tracking " + err.Error())
	}

	linesChan := t.Lines()

	for {
		select {
		case s := <-linesChan:
			ev, err := fr.parser.Parse(string(s.Bytes()))
			if err != nil {
				if err != io.EOF && err.Error() != "EOF" {
					fr.errorChan <- err
				}
				continue // TODO wrong format is not considered an error, I'll handle this better later
			}

			if ev == nil {
				continue
			}

			for _, opt := range fr.options {
				*ev, err = opt.Apply(*ev)
				if err != nil {
					fr.errorChan <- err
					continue
				}
			}

			fr.eventChan <- *ev

		case <-fr.done:
			close(linesChan)
			return
		}
	}
}
