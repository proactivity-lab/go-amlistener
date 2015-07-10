// Author  Raido Pahtma
// License MIT

package main

import "fmt"
import "os"
import "os/signal"
import "log"
import "time"

import "github.com/jessevdk/go-flags"
import "github.com/proactivity-lab/go-sfconnection"

const ApplicationVersionMajor = 0
const ApplicationVersionMinor = 1
const ApplicationVersionPatch = 2

var ApplicationBuildDate string
var ApplicationBuildDistro string

func main() {

	var opts struct {
		Positional struct {
			ConnectionString string `description:"Connectionstring sf@HOST:PORT"`
		} `positional-args:"yes"`

		Reconnect uint `long:"reconnect" default:"30" description:"Reconnect period, seconds"`

		Debug       []bool `short:"D" long:"debug" description:"Debug mode, print raw packets"`
		ShowVersion func() `short:"V" long:"version" description:"Show application version"`
	}

	opts.ShowVersion = func() {
		if ApplicationBuildDate == "" {
			ApplicationBuildDate = "YYYY-mm-dd_HH:MM:SS"
		}
		if ApplicationBuildDistro == "" {
			ApplicationBuildDistro = "unknown"
		}
		fmt.Printf("amlistener %d.%d.%d (%s %s)\n", ApplicationVersionMajor, ApplicationVersionMinor, ApplicationVersionPatch, ApplicationBuildDate, ApplicationBuildDistro)
		os.Exit(0)
	}

	_, err := flags.Parse(&opts)
	if err != nil {
		fmt.Printf("Argument parser error: %s\n", err)
		os.Exit(1)
	}

	host, port, err := sfconnection.ParseSfConnectionString(opts.Positional.ConnectionString)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	dsp := sfconnection.NewMessageDispatcher(sfconnection.NewMessage(0, 0))
	receive := make(chan *sfconnection.Message)
	dsp.RegisterMessageSnooper(receive)

	sfc := sfconnection.NewSfConnection()
	sfc.AddDispatcher(dsp)

	// Configure logging
	logformat := log.Ldate | log.Ltime | log.Lmicroseconds
	var logger *log.Logger
	if len(opts.Debug) > 0 {
		if len(opts.Debug) > 1 {
			logformat = logformat | log.Lshortfile
		}
		logger = log.New(os.Stdout, "INFO:  ", logformat)
		sfc.SetDebugLogger(log.New(os.Stdout, "DEBUG: ", logformat))
		sfc.SetInfoLogger(logger)
	} else {
		logger = log.New(os.Stdout, "", logformat)
	}
	sfc.SetWarningLogger(log.New(os.Stdout, "WARN:  ", logformat))
	sfc.SetErrorLogger(log.New(os.Stdout, "ERROR: ", logformat))

	// Connect to the host
	sfc.Autoconnect(host, port, time.Duration(opts.Reconnect)*time.Second)

	// Set up signals to close nicely on Control+C
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, os.Kill)

	for interrupted := false; interrupted == false; {
		select {
		case msg := <-receive:
			logger.Printf("%s\n", msg)
		case sig := <-signals:
			signal.Stop(signals)
			logger.Printf("signal %s\n", sig)
			sfc.Disconnect()
			interrupted = true
		}
	}
}
