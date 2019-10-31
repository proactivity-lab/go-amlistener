// Author  Raido Pahtma
// License MIT

package main

import "fmt"
import "os"
import "os/signal"
import "log"
import "time"

import "github.com/jessevdk/go-flags"
import "github.com/proactivity-lab/go-moteconnection"

const ApplicationVersionMajor = 0
const ApplicationVersionMinor = 2
const ApplicationVersionPatch = 1

var ApplicationBuildDate string
var ApplicationBuildDistro string

func main() {

	var opts struct {
		Positional struct {
			ConnectionString string `description:"Connectionstring sf@HOST:PORT or serial@PORT:BAUD"`
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
		flagserr := err.(*flags.Error)
		if flagserr.Type != flags.ErrHelp {
			if len(opts.Debug) > 0 {
				fmt.Printf("Argument parser error: %s\n", err)
			}
			os.Exit(1)
		}
		os.Exit(0)
	}

	conn, cs, err := moteconnection.CreateConnection(opts.Positional.ConnectionString)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	dsp := moteconnection.NewMessageDispatcher(moteconnection.NewMessage(0, 0))
	receive := make(chan moteconnection.Packet)
	dsp.RegisterMessageSnooper(receive)

	conn.AddDispatcher(dsp)

	// Configure logging
	logformat := log.Ldate | log.Ltime | log.Lmicroseconds
	var logger *log.Logger
	if len(opts.Debug) > 0 {
		if len(opts.Debug) > 1 {
			logformat = logformat | log.Lshortfile
		}
		logger = log.New(os.Stdout, "INFO:  ", logformat)
		conn.SetDebugLogger(log.New(os.Stdout, "DEBUG: ", logformat))
		conn.SetInfoLogger(logger)
	} else {
		logger = log.New(os.Stdout, "", logformat)
	}
	conn.SetWarningLogger(log.New(os.Stdout, "WARN:  ", logformat))
	conn.SetErrorLogger(log.New(os.Stdout, "ERROR: ", logformat))

	// Connect to the host
	logger.Printf("Connecting to %s\n", cs)
	conn.Autoconnect(time.Duration(opts.Reconnect) * time.Second)

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
			conn.Disconnect()
			interrupted = true
		}
	}
}
