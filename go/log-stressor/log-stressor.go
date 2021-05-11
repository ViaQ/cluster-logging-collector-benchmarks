package main

import (
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

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

type Options struct {
	payloadGen        string
	distribution      string
	payloadSize       int
	messagesPerSecond int
	totMessages       int
	outputFormat      string
	outputFile        string
}

func main() {

	opt := Options{}

	flag.StringVar(&opt.outputFile, "file", "", "The file to output (default: STDOUT)")
	flag.StringVar(&opt.outputFormat, "output-format", "default", "The output format: default, crio (mimic CRIO output)")
	flag.StringVar(&opt.payloadGen, "payload-gen", "constant", "Payload generator [enum]: constant(default), fixed")
	flag.StringVar(&opt.distribution, "distribution", "fixed", "Payload distribution [enum] (default = fixed)")
	flag.IntVar(&opt.payloadSize, "payload_size", 100, "Payload length [int] (default = 100)")
	flag.IntVar(&opt.messagesPerSecond, "msgpersec", 1, "Number of messages per second (default = 1)")
	flag.IntVar(&opt.totMessages, "totMessages", 1, "Total number of messages (only applicable for 'fixed' payload-gen")

	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
	hash := fmt.Sprintf("%032X", rnd.Uint64())

	outFormat := NewFormatter(opt.outputFile, opt.outputFormat)
	generateLogs := NewDistributionProfile(opt.payloadGen, opt.distribution)
	generateLogs(outFormat, hash, opt)
}

type DistributionProfile func(format formatter, hash string, opt Options)

//FixedFixedProfile produces a fixed number of messages with a constant distribution?
func FixedFixedProfile(format formatter, hash string, opt Options) {
	for i := 0; i < opt.totMessages; i++ {
		payload := RandStringBytes(opt.payloadSize)
		format(hash, i, payload)
	}
}

//ConstantFixedProfile produces a constant stream of fixed messages
func ConstantFixedProfile(format formatter, hash string, opt Options) {
	bursts := 1
	if opt.messagesPerSecond > minBurstMessageCount {
		bursts = numberOfBursts
	}
	messageCount := 0
	startTime := time.Now().Unix() - 1
	for {
		for i := 0; i < opt.messagesPerSecond/bursts; i++ {
			payload := RandStringBytes(opt.payloadSize)
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

func NewDistributionProfile(payloadGen, distribution string) DistributionProfile {
	profile := payloadGen + "/" + distribution
	switch profile {
	case "fixed/fixed":
		return FixedFixedProfile
	default:
		return ConstantFixedProfile
	}
}

type formatter func(hash string, messageCount int, payload string)

func NewFormatter(outputFile, outputFormat string) formatter {
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			log.Fatalf("Unable to create out file %s: %v", outputFile, err)
		}
		log.SetOutput(f)
	}
	formatter := FormatForStdOut
	if outputFormat == "crio" {
		log.SetFlags(0)
		log.SetPrefix("")
		formatter = FormatForCrio
	}
	return formatter
}

func FormatForCrio(hash string, messageCount int, payload string) {
	now := time.Now().Format(time.RFC3339Nano)
	log.Printf("%s stdout F goloader seq - %s - %010d - %s\n", now, hash, messageCount, payload)
}

func FormatForStdOut(hash string, messageCount int, payload string) {
	log.Printf("goloader seq - %s - %010d - %s", hash, messageCount, payload)
}
