package main

import (
	"flag"
	logger "github.com/ViaQ/cluster-logging-load-client/loadclient"
	log "github.com/sirupsen/logrus"
)

type options struct {
	payloadGen        string
	distribution      string
	payloadSize       int
	messagesPerSecond int
	totMessages       int
	outputFormat      string
	outputFile        string
	useLogSamples     string
}

func main() {

	opt := options{}

	flag.StringVar(&opt.outputFile, "file", "", "The file to output (default: STDOUT)")
	flag.StringVar(&opt.outputFormat, "output-format", "default", "The output format: default, crio (mimic CRIO output)")
	flag.StringVar(&opt.payloadGen, "payload-gen", "constant", "Payload generator [enum]: constant(default), fixed")
	flag.StringVar(&opt.distribution, "distribution", "fixed", "Payload distribution [enum] (default = fixed)")
	flag.IntVar(&opt.payloadSize, "payload_size", 100, "Payload length [int] (default = 100)")
	flag.IntVar(&opt.messagesPerSecond, "msgpersec", 1, "Number of messages per second (default = 1)")
	flag.IntVar(&opt.totMessages, "totMessages", 1, "Total number of messages (only applicable for 'fixed' payload-gen")
	flag.StringVar(&opt.useLogSamples, "use_log_samples", "false", "Use log samples or not [enum] (default = false)")

	flag.Parse()

	// define "destination" as file or stdout
	totalMessages := int64(opt.totMessages)
	if opt.payloadGen == "constant" {
		totalMessages = 0
	}

	// define "destination" as file or stdout
	destination := "stdout"
	if opt.outputFile != "" {
		destination = "file"
	}

	source := "synthetic"
	if opt.useLogSamples == "true" {
		source = "application"
	}

	loggerOptions := logger.Options{
		Command:              logger.Generate,
		Threads:              1,
		LogLinesPerSec:       int64(opt.messagesPerSecond),
		Destination:          destination,
		Source:               source,
		SyntheticPayloadSize: opt.payloadSize,
		TotalLogLines:        totalMessages,
		LogFormat:            opt.outputFormat,
		OutputFile:           opt.outputFile,
		DestinationAPIURL:    "",
	}
	log.SetLevel(log.ErrorLevel)
	logger.GenerateLog(loggerOptions)
}