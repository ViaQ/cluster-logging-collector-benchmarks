package main

import (
    "flag"
    "fmt"
    "log"
    "math/rand"
    "strings"
    "time"
)

const  minBurstMessageCount = 100
const  numberOfBursts = 10

func main() {

    var payloadGen string
    var distribution string
    var payloadSize int
    var messagesPerSecond int

    flag.StringVar(&payloadGen, "payload-gen", "fixed", "Payload generator [enum] (default = fixed)")
    flag.StringVar(&distribution, "distribution", "fixed", "Payload distribution [enum] (default = fixed)")
    flag.IntVar(&payloadSize, "payload_size", 10, "Payload length [int] (default = 10)")
    flag.IntVar(&messagesPerSecond, "msgpersec", 1, "Number of messages per second (default = 1)")

    flag.Parse()

    var rnd = rand.New( rand.NewSource(time.Now().UnixNano()))
    hash := fmt.Sprintf("%032X", rnd.Uint64())

    bursts := 1
    if messagesPerSecond > minBurstMessageCount {
        bursts = numberOfBursts
    }

    messageCount := 0
    startTime := time.Now().Unix() -1
    for {
        payload := strings.Repeat("*", payloadSize)
        for i := 0; i < messagesPerSecond/bursts; i++ {
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