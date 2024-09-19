package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	DEBUG bool

	filename       = flag.String("f", "/var/log/auth.log", "ssh log file source")
	prometheusPort = flag.String("m", ":2112", "prometheus port")
	geolocate      = flag.Bool("g", true, "geolocation service enabled/disabled")
	debug          = flag.Bool("d", false, "debug mode enabled")
)

const (
	NL = "\n"

	apiStackKey = "SSH2PROM_IPSTACK_ACCESSKEY"
)

func log(msg interface{}, args ...interface{}) {
	t := time.Now().Local()
	ts := fmt.Sprintf(
		"%03d."+"%02d%02d."+"%02d"+"%02d.",
		t.Year()%1000, t.Month(), t.Day(), t.Hour(), t.Minute(),
	)
	msgtext := fmt.Sprintf("%s %s", ts, msg) + NL
	fmt.Fprintf(os.Stderr, msgtext, args...)
}

func main() {
	flag.Parse()

	log("starting up")

	if *debug {
		DEBUG = true
		log("DEBUG messages will be logged")
	}

	// Setup geolocation services
	geolocationServices := []Geolocator{ipAPI{}}

	apiStackAccessKey := os.Getenv(apiStackKey)
	if apiStackAccessKey != "" {
		geolocationServices = append(geolocationServices, apiStack{AccessKey: apiStackAccessKey})
	}

	geolocator := NewGeolocationProvider(geolocationServices...)
	locatorOpt := geolocateOption{geolocator}

	// Setup parser
	parser := NewFailedConnEventParser()

	// Setup reader
	readerOpts := []ReaderOption{}
	if *geolocate {
		readerOpts = append(readerOpts, locatorOpt)
	}

	respChan := make(chan FailedConnEvent, 100)
	errorChan := make(chan error, 100)

	reader := NewFileReader(*filename, parser, respChan, errorChan, readerOpts...)
	go reader.Start()
	log("started reader for file `%s`", *filename)
	defer reader.Stop()
	defer close(respChan)
	defer close(errorChan)

	// Setup prometheus reporter
	rep := prometheusReporter{}
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(*prometheusPort, nil)

	// Setup shutdown from OS signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {

		case ev := <-respChan:
			rep.Report(ev)
			if DEBUG {
				log("DEBUG reported %v", ev)
			}

		case err := <-errorChan:
			log("ERROR %#v", err)

		case _ = <-sigs:
			log("shutting down")
			os.Exit(0)

		}
	}
}
