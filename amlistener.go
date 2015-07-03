// Author  Raido Pahtma
// License MIT

package main

import "fmt"
import "os"
import "os/signal"
import "log"
import "time"
import "regexp"
import "strconv"

import "github.com/jessevdk/go-flags"
import "github.com/proactivity-lab/go-sf-connection"

const ApplicationVersionMajor = 0
const ApplicationVersionMinor = 1
const ApplicationVersionPatch = 0

var ApplicationBuildDate string
var ApplicationBuildDistro string

func main() {

	var opts struct {
		Positional struct {
			ConnectionString string `description:"Connectionstring sf@HOST:PORT"`
		} `positional-args:"yes"`
		Reconnect   uint   `long:"reconnect" default:"30" description:"Reconnect period, seconds"`
		Debug       bool   `short:"D" long:"debug" default:"false" description:"Debug mode, print raw packets"`
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
		fmt.Printf("Argument parser error: %s", err)
		os.Exit(1)
	}

	host := "localhost"
	port := 9002

	if opts.Positional.ConnectionString != "" {
		re := regexp.MustCompile("sf@([a-zA-Z0-9.-]+)(:([0-9]+))?")
		match := re.FindStringSubmatch(opts.Positional.ConnectionString) // [sf@localhost:9002 localhost :9002 9002]
		if len(match) == 4 {
			host = match[1]
			if len(match[3]) > 0 {
				p, err := strconv.ParseUint(match[3], 10, 16)
				if err == nil {
					port = int(p)
				} else {
					fmt.Printf("ERROR: %s cannot be used as a TCP port number!\n", match[2])
					os.Exit(1)
				}
			}
		} else {
			fmt.Printf("ERROR: %s cannot be used as a connectionstring!\n", opts.Positional.ConnectionString)
			os.Exit(1)
		}
	}

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, os.Kill)

	dsp := sfconnection.NewPacketDispatcher(0, new(sfconnection.Message))
	receive := make(chan sfconnection.Packet)
	dsp.RegisterSnooper(receive)

	sfc := sfconnection.NewSfConnection()
	sfc.AddDispatcher(dsp)

	logformat := log.Ldate | log.Ltime | log.Lmicroseconds
	var logger *log.Logger
	if opts.Debug {
		logger = log.New(os.Stdout, "INFO:  ", logformat)
		sfc.SetDebugLogger(log.New(os.Stdout, "DEBUG: ", logformat))
		sfc.SetInfoLogger(logger)
	} else {
		logger = log.New(os.Stdout, "", logformat)
	}
	sfc.SetWarningLogger(log.New(os.Stdout, "WARN:  ", logformat))
	sfc.SetErrorLogger(log.New(os.Stdout, "ERROR: ", logformat))

	sfc.Autoconnect(host, port, time.Duration(opts.Reconnect)*time.Second)

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
