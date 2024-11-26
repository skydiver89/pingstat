package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"time"

	"github.com/go-ping/ping"
)

const VERSION = "0.1.0"

var count int
var interval int //milliseconds
var timeout int  //seconds
var ipAddress string

func parseFlags() {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [-c packets_count] [-i interval(msec)] [-t timeout(sec)] [-a address] [-h] [-v]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}
	flag.StringVar(&ipAddress, "a", "8.8.8.8", "Address to ping")
	flag.IntVar(&count, "c", 100, "Number of packets per interval")
	flag.IntVar(&interval, "i", 90, "Interval between pings in milliseconds")
	flag.IntVar(&timeout, "t", 10, "Timeout for interval in seconds")
	var showHelp bool
	flag.BoolVar(&showHelp, "h", false, "Show this help")
	var showVersion bool
	flag.BoolVar(&showVersion, "v", false, "Show version")
	flag.Parse()
	if showHelp {
		flag.Usage()
		os.Exit(0)
	}
	if showVersion {
		fmt.Printf("%s version: %s\n", os.Args[0], VERSION)
		os.Exit(0)
	}
}

func createPinger() *ping.Pinger {
	pinger, err := ping.NewPinger(ipAddress)
	if err != nil {
		log.Fatalf("Error creating pinger: %v\n", err)
	}
	pinger.SetPrivileged(true)
	pinger.Count = count
	pinger.Interval = time.Millisecond * time.Duration(interval)
	pinger.Timeout = time.Second * time.Duration(timeout)
	return pinger
}

func main() {
	parseFlags()
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("Unable to get current user: %s\n", err)
	}
	if currentUser.Username != "root" {
		log.Fatalln("This program should be run as root")
	}

	logFile, err := os.Create("ping_stats.log")
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)
	logger.Printf("Address: %s Count per interval: %d Interval(msec): %d Timeout(sec): %d\n", ipAddress, count, interval, timeout)
	log.Printf("Address: %s Count per interval: %d Interval(msec): %d Timeout(sec): %d\n", ipAddress, count, interval, timeout)

	for {
		pinger := createPinger()
		err = pinger.Run()
		if err != nil {
			log.Printf("Error running pinger: %v", err)
			continue
		}
		stats := pinger.Statistics()
		logMessage := fmt.Sprintf(
			"Sent: %3d "+
				"Received: %3d "+
				"Loss: %.2f%% "+
				"Min: %d "+
				"Max: %d "+
				"Avg: %d ",
			stats.PacketsSent,
			stats.PacketsRecv,
			stats.PacketLoss,
			stats.MinRtt/time.Millisecond,
			stats.MaxRtt/time.Millisecond,
			stats.AvgRtt/time.Millisecond,
		)
		logger.Println(logMessage)
		log.Println(logMessage)
	}
}
