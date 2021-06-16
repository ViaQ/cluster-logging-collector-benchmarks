package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

const (
	minBurstMessageCount = 100
	numberOfBursts       = 10
	letterBytes          = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func readFile(fileName string)[] string {

    readFile, err := os.Open(fileName)

	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	var fileTextLines []string

	for fileScanner.Scan() {
		fileTextLines = append(fileTextLines, fileScanner.Text())
	}

	_ = readFile.Close()
    return fileTextLines
}

func getLogLinesFromFile()[]string{
	var loglines []string
	fileName := "samples.log"
	loglines = readFile(fileName)
	return loglines
}

func getRandomLogline(loglines []string) string{
    index := rand.Intn(len(loglines))
    return loglines[index]
}

func getPayload(opt options, loglines []string) string{
	payload := ""
	if opt.useLogSamples == "true" {
		payload = getRandomLogline(loglines)
	}else{
		payload = randStringBytes(opt.payloadSize)
	}
	return payload
}

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

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

	rand.Seed(time.Now().UnixNano())

	var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
	hash := fmt.Sprintf("%032X", rnd.Uint64())

	outFormat := newFormatter(opt.outputFile, opt.outputFormat)
	generateLogs := newDistributionProfile(opt.payloadGen, opt.distribution)
	generateLogs(outFormat, hash, opt)
}

type distributionProfile func(format formatter, hash string, opt options)

//finiteFixedProfile produces a fixed number of messages with a constant distribution?
func finiteFixedProfile(format formatter, hash string, opt options) {
    
	loglines := getLogLinesFromFile()

	for i := 0; i < opt.totMessages; i++ {
		payload := getPayload(opt, loglines)
		format(hash, i, payload)
	}
}

//constantFixedProfile produces a constant stream of fixed messages
func constantFixedProfile(format formatter, hash string, opt options) {
	loglines := getLogLinesFromFile()

	bursts := 1
	if opt.messagesPerSecond > minBurstMessageCount {
		bursts = numberOfBursts
	}
	messageCount := 0
	startTime := time.Now().Unix() - 1
	for {
		for i := 0; i < opt.messagesPerSecond/bursts; i++ {
			payload := getPayload(opt, loglines)
			format(hash, messageCount, payload)
			messageCount++
		}

		sleep := 1.0 / float64(bursts)
		deltaTime := int(time.Now().Unix() - startTime)

		messagesLoggedPerSec := messageCount / deltaTime
		if messagesLoggedPerSec >= opt.messagesPerSecond {
			time.Sleep(time.Duration(sleep * float64(time.Second)))
		}
	}
}

func newDistributionProfile(payloadGen, distribution string) distributionProfile {
	profile := payloadGen + "/" + distribution
	switch profile {
	case "fixed/fixed":
		return finiteFixedProfile
	default:
		return constantFixedProfile
	}
}

type formatter func(hash string, messageCount int, payload string)

func newFormatter(outputFile, outputFormat string) formatter {
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			log.Fatalf("Unable to create out file %s: %v", outputFile, err)
		}
		log.SetOutput(f)
	}
	formatter := formatForStdOut
	if outputFormat == "crio" {
		log.SetFlags(0)
		log.SetPrefix("")
		formatter = formatForCrio
	}
	return formatter
}

func formatForCrio(hash string, messageCount int, payload string) {
	now := time.Now().Format(time.RFC3339Nano)
	log.Printf("%s stdout F goloader seq - %s - %010d - %s\n", now, hash, messageCount, payload)
}

func formatForStdOut(hash string, messageCount int, payload string) {
	log.Printf("goloader seq - %s - %010d - %s", hash, messageCount, payload)
}
