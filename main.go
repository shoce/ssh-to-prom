// GoGet GoFmt GoBuildNull

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	promhttp "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	DEBUG bool

	filename       = flag.String("f", "/var/log/auth.log", "ssh log file source")
	prometheusPort = flag.String("m", ":2112", "prometheus port")
	debug          = flag.Bool("d", false, "debug mode enabled")
)

const (
	SP = " "
	NL = "\n"
)

func perr(msg string, args ...interface{}) {
	tnow := time.Now().Local()
	ts := fmt.Sprintf(
		"%03d%02d%02d.%02d%02d",
		tnow.Year()%1000, tnow.Month(), tnow.Day(), tnow.Hour(), tnow.Minute(),
	)
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, ts+SP+msg+NL)
	} else {
		fmt.Fprintf(os.Stderr, ts+SP+msg+NL, args...)
	}
}

func main() {
	flag.Parse()

	perr("starting up")

	if *debug {
		DEBUG = true
		perr("DEBUG")
	}

	parser := ConnEventParser{}

	readerOpts := []ReaderOption{}

	eventChan := make(chan ConnEvent, 1000)
	errorChan := make(chan error, 100)

	reader := NewFileReader(*filename, parser, eventChan, errorChan, readerOpts...)
	go reader.Start()
	perr("started reader for file [%s]", *filename)
	defer reader.Stop()
	defer close(eventChan)
	defer close(errorChan)

	rep := prometheusReporter{}
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(*prometheusPort, nil)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {

		case ev := <-eventChan:
			rep.Report(ev)
			if DEBUG {
				perr(
					"DEBUG reported @{ @Timestamp [%s] @Accepted <%t> @AuthMethod [%s] @User [%s] @Addr [%s] @Port <%d> }",
					ev.Timestamp.Format("20060102.150405"), ev.Accepted, ev.AuthMethod, ev.User, ev.Addr, ev.Port,
				)
			}

		case err := <-errorChan:
			perr("ERROR %#v", err)

		case _ = <-sigs:
			perr("shutting down")
			os.Exit(0)

		}
	}
}
