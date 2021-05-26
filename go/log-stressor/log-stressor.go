package main

import (
    "flag"
    "fmt"
    "log"
    "math/rand"
    "time"
    "bufio"
    "os"
)

const  minBurstMessageCount = 100
const  numberOfBursts = 10

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func ReadFile(fileName string)[] string {

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

	readFile.Close()
    return fileTextLines
}

func RandRecord(loglines []string) string{
    index := rand.Intn(len(loglines))
    return loglines[index]
}

func RandStringBytes(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = letterBytes[rand.Intn(len(letterBytes))]
    }
    return string(b)
}

func main() {

    var payloadGen string
    var distribution string
    var payloadSize int
    var messagesPerSecond int
    var datafile string
    var loglines []string
    var fileName string
    fileName = "samples.log"
    useLogs := false

    flag.StringVar(&payloadGen, "payload-gen", "fixed", "Payload generator [enum] (default = fixed)")
    flag.StringVar(&distribution, "distribution", "fixed", "Payload distribution [enum] (default = fixed)")
    flag.IntVar(&payloadSize, "payload_size", 100, "Payload length [int] (default = 100)")
    flag.IntVar(&messagesPerSecond, "msgpersec", 1, "Number of messages per second (default = 1)")
    flag.StringVar(&datafile, "datafile", "false", "Use data file or not [enum] (default = false)")

    flag.Parse()

    if datafile == "true" {
        useLogs = true
    }

    if useLogs == true {
        loglines = ReadFile(fileName)
    }

    rand.Seed(time.Now().UnixNano())

    var rnd = rand.New( rand.NewSource(time.Now().UnixNano()))
    hash := fmt.Sprintf("%032X", rnd.Uint64())
    bursts := 1
    if messagesPerSecond > minBurstMessageCount {
        bursts = numberOfBursts
    }
    log.SetFlags(0)
    messageCount := 0
    startTime := time.Now().Unix() -1
    payload := ""
    for {
        for i := 0; i < messagesPerSecond/bursts; i++ {

            if useLogs == true {
                payload = RandRecord(loglines)
            }else{
                payload = RandStringBytes(payloadSize)
            }
            log.Printf("goloader seq - %s - %010d - %s",hash, messageCount, payload)
            messageCount ++
        }
        
        sleep := 1.0/ float64(bursts)
        deltaTime := int(time.Now().Unix() - startTime)

        messagesLoggedPerSec :=  messageCount / deltaTime
        if messagesLoggedPerSec >= messagesPerSecond {
            time.Sleep(time.Duration(sleep * float64(time.Second)))
        }
    }
}