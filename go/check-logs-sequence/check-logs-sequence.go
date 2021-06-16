package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"encoding/json"

	"github.com/papertrail/go-tail/follower"
	log "github.com/sirupsen/logrus"
)

type logSourceInfo struct {
	loggedCount    int64
	collectedCount int64
	smallestSeq    int64
	biggestSeq     int64
	lastCollected  int64
	hashID         string
}

type reportStatistics struct {
	totalLogsCollectedCount int64
	totalLogsSkippedCount   int64
	startMonitoringTime     time.Time
}

func main() {

	var fluentLogFileName string
	var reportCount int64
	var logLevel string
	var reportFormat string
	var seekFrom int

	flag.StringVar(&fluentLogFileName, "f", "0.log", "fluent log file to tail")
	flag.Int64Var(&reportCount, "c", 100, "number of logs between reports")
	flag.StringVar(&logLevel, "l", "fatal", "Logging level e.g debug, info")
	flag.StringVar(&reportFormat, "rf", "default", "Report format: ndjson, default")
	flag.IntVar(&seekFrom, "s", io.SeekEnd, "Tail seek from: 2 - end , 1 - current, 0 - start")
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	lvl, err := log.ParseLevel(logLevel)
	if err != nil {
		lvl = log.FatalLevel
	}
	log.SetLevel(lvl)

	report := defaultReporter
	if *&reportFormat == "ndjson" {
		report = ndjsonReporter
	}

	logsCurrentInfo := make(map[string]logSourceInfo)
	logsTotalInfo := make(map[string]logSourceInfo)
	reportData := reportStatistics{}

	reportData.startMonitoringTime = time.Now().Add(-1 * time.Second)

	// forever loop tailing on the file
	for {
		log.Info("Watching file: ", fluentLogFileName)
		time.Sleep(1 * time.Millisecond)

		t, err := follower.New(fluentLogFileName, follower.Config{
			Whence: seekFrom,
			Offset: 0,
			Reopen: true,
		})

		if err != nil {
			log.Error("follower.New: Error ", err)
			continue
		}

		for line := range t.Lines() {
			name, seq, logTag, hashID, errParse := parseLine(line.String())
			if errParse != nil {
				log.Error("Error in Line: ", line.String())
				log.Error(errParse)
				reportData.totalLogsSkippedCount++
				continue
			}

			// Skip not full log lines
			if logTag != "F" {
				log.Error("Skipping line with logTag not [F]: ", line.String())
				reportData.totalLogsSkippedCount++
				continue
			}

			if _, ok := logsCurrentInfo[name]; !ok {
				logsCurrentInfo[name] = logSourceInfo{hashID: hashID, smallestSeq: seq, biggestSeq: seq}
			}

			if _, ok := logsTotalInfo[name]; !ok {
				logsTotalInfo[name] = logSourceInfo{hashID: hashID, smallestSeq: seq, biggestSeq: seq}
			}

			// Get current entry information
			entry := logsCurrentInfo[name]
			if entry.hashID != hashID {
				fmt.Printf("Error in Line: %s\n", line.String())
				fmt.Printf("Source Identification Hash ID (current) is wrong for [%s]: should be %s but is %s\n", name, entry.hashID, hashID)
				reportData.totalLogsSkippedCount++
				continue
			}

			// Get total entry information
			totalEntry := logsTotalInfo[name]
			if totalEntry.hashID != hashID {
				fmt.Printf("Error in Line: %s\n", line.String())
				fmt.Printf("Source Identification Hash ID (total) is wrong for [%s]: should be %s but is %s\n", name, totalEntry.hashID, hashID)
				reportData.totalLogsSkippedCount++
				continue
			}

			// calculate current metrics
			if entry.lastCollected != 0 && seq < entry.lastCollected {
				log.Error("Error in Line: ", line.String())
				log.Errorf("Out of order sequence for[%s]: last collected seq is %d but we got %d\n", name, entry.lastCollected, seq)
			}
			if seq < entry.smallestSeq {
				entry.smallestSeq = seq
			}

			if seq > entry.biggestSeq {
				entry.biggestSeq = seq
			}

			entry.collectedCount++
			entry.lastCollected = seq
			entry.loggedCount = entry.biggestSeq - entry.smallestSeq + 1

			// calculate total metrics
			if seq < totalEntry.smallestSeq {
				totalEntry.smallestSeq = seq
			}
			if seq > totalEntry.biggestSeq {
				totalEntry.biggestSeq = seq
			}

			totalEntry.collectedCount++
			totalEntry.loggedCount = totalEntry.biggestSeq - totalEntry.smallestSeq + 1

			// calculate global metrics
			reportData.totalLogsCollectedCount++

			// persist
			logsCurrentInfo[name] = entry
			logsTotalInfo[name] = totalEntry

			if reportData.totalLogsCollectedCount%reportCount == 0 {
				report(reportData, logsCurrentInfo, logsTotalInfo)

				// reset counting
				logsCurrentInfo = make(map[string]logSourceInfo)
			}
		}
	}
}
func parseLine(line string) (name string, seq int64, logTag string, hashID string, err error) {
	// Parse logTag
	logTagStartIndex := strings.Index(line, "\"logtag\":\"")
	if logTagStartIndex > 0 {
		logTagStartIndex += len("\"logtag\":\"")
		logTag = line[logTagStartIndex : logTagStartIndex+1]
	} else {
		logTag = "F" // assuming that if we do not have log tag it is a full line
	}

	// parse name (from path)
	pathStartIndex := strings.Index(line, "\"path\":\"")
	if pathStartIndex == -1 {
		err = errors.New("parseLine: cant find path start")
		return  "", 0, "", "", err
	}
	pathStartIndex += len("\"path\":\"")

	pathEndIndex := strings.Index(line[pathStartIndex:], "\"")
	if pathEndIndex == -1 {
		err = errors.New("parseLine: cant find path end")
		return  "", 0, "", "", err
	}

	path := line[pathStartIndex : pathStartIndex+pathEndIndex]

	//// get container name from path
	nameSliced := strings.Split(path, "_")
	if len(nameSliced) < 1 {
		err = errors.New("parseLine: can't parse _ in path")
		return  "", 0, "", "", err
	}
	nameSliced = strings.Split(fmt.Sprintf("%s", nameSliced[0]), "/")
	if len(nameSliced) < 5 {
		err = errors.New("parseLine: can't parse / in path")
		return  "", 0, "", "", err
	}

	if nameSliced[3] != "containers" {
		err = errors.New("parseLine: can't parse / path -> follow only  /var/log/containers ")
		return  "", 0, "", "", err
	}
	name = nameSliced[4]

	// parse sequence and hashID (from message)
	messageStartIndex := strings.Index(line, "\"message\":\"")
	if messageStartIndex == -1 {
		err = errors.New("parseLine: cant find message start")
		return  "", 0, "", "", err
	}
	messageStartIndex += len("\"message\":\"")

	messageEndIndex := strings.Index(line[messageStartIndex:], "\"")
	if messageEndIndex == -1 {
		err = errors.New("parseLine: cant find message end")
		return  "", 0, "", "", err
	}
	message := line[messageStartIndex : messageStartIndex+messageEndIndex]

	// get the sequence number of the log
	logSliced := strings.Split(message, "-")
	if len(logSliced) < 3 {
		err = errors.New("parseLine: can't parse - in log")
		return  "", 0, "", "", err
	}
	seqStr := strings.TrimSpace(logSliced[2])
	seq, err = strconv.ParseInt(seqStr, 10, 0)
	if err != nil {
		err = errors.New("parseLine: can't parse ParseInt in log")
		return  "", 0, "", "", err
	}

	// get the hashID from the log
	hashID = strings.TrimSpace(logSliced[1])
	return name, seq, logTag, hashID, nil
}
func ndjsonReporter(reportData reportStatistics, logsCurrentInfo map[string]logSourceInfo, logsTotalInfo map[string]logSourceInfo) {
	now := time.Now()
	deltaTimeInSeconds := now.Unix() - reportData.startMonitoringTime.Unix()
	summary := map[string]interface{}{}
	summary["time"] = now.Format(time.RFC3339)
	if !reportData.startMonitoringTime.IsZero() {
		summary["timeFromStartMonitoringSec"] = deltaTimeInSeconds
		summary["totLogsCollected"] = reportData.totalLogsCollectedCount
		summary["totLogsCollectedPerSec"] = reportData.totalLogsCollectedCount / deltaTimeInSeconds
		summary["totalLogsSkipped"] = reportData.totalLogsSkippedCount
	}

	var apps []interface{}
	for name, totalEntry := range logsTotalInfo {
		pod := map[string]interface{}{}
		entry := logsCurrentInfo[name]
		pod["container"] = name
		pod["currentLogged"] = entry.loggedCount
		pod["currentCollected"] = entry.collectedCount
		pod["currentLoss"] = entry.loggedCount - entry.collectedCount
		pod["totalLoggedPerSec"] = totalEntry.loggedCount / deltaTimeInSeconds
		pod["totalLogged"] = totalEntry.loggedCount
		pod["totalCollected"] = totalEntry.collectedCount
		pod["totalCollectedPerSec"] = totalEntry.collectedCount / deltaTimeInSeconds
		pod["totalLoss"] = totalEntry.loggedCount - totalEntry.collectedCount

		apps = append(apps, pod)
	}

	stats := map[string]interface{}{
		"apps":    apps,
		"summary": summary,
	}
	ndjson, err := json.Marshal(stats)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%s\n", ndjson)
}

func defaultReporter(reportData reportStatistics, logsCurrentInfo map[string]logSourceInfo, logsTotalInfo map[string]logSourceInfo) {
	now := time.Now()
	deltaTimeInSeconds := now.Unix() - reportData.startMonitoringTime.Unix()
	fmt.Printf("Report at: %s\n", now.String())
	fmt.Printf("-==-=-=-=-=\n")
	if !reportData.startMonitoringTime.IsZero() {
		fmt.Printf("Time from start monitoring (in secs): %d\n", deltaTimeInSeconds)
		fmt.Printf("Total number of collected logs: %d\n", reportData.totalLogsCollectedCount)
		fmt.Printf("Total collected logs per sec: %d\n", reportData.totalLogsCollectedCount/deltaTimeInSeconds)
		fmt.Printf("Skipped log lines: %d\n", reportData.totalLogsSkippedCount)
	}

	fmt.Printf("-==-=-=-=-=\n")
	tableFormat := "| %-36v | %-9v | %-9v | %-9v | %-9v | %-9v | %-9v | %-9v | %-9v |\n"
	tableFormatLen := len(fmt.Sprintf(tableFormat, 0, 0, 0, 0, 0, 0, 0, 0, 0)) - 1
	fmt.Printf(strings.Repeat("-", tableFormatLen) + "\n")
	fmt.Printf(tableFormat,
		"",
		"Current",
		"Lines",
		"",
		"Total",
		"Lines",
		"",
		"",
		"")
	fmt.Printf(strings.Repeat("-", tableFormatLen) + "\n")
	fmt.Printf(tableFormat,
		"Container name",
		"Logged",
		"Collected",
		"Loss",
		"Logged",
		"Lo./Sec",
		"Collected",
		"Co./Sec",
		"Loss",
	)
	fmt.Printf(strings.Repeat("-", tableFormatLen) + "\n")

	names := make([]string, 0)
	for k := range logsTotalInfo {
		names = append(names, k)
	}
	sort.Strings(names)

	for _, name := range names {
		entry := logsCurrentInfo[name]
		totalEntry := logsTotalInfo[name]
		fmt.Printf(tableFormat, name,
			entry.loggedCount,
			entry.collectedCount,
			entry.loggedCount-entry.collectedCount,
			totalEntry.loggedCount,
			totalEntry.loggedCount/deltaTimeInSeconds,
			totalEntry.collectedCount,
			totalEntry.collectedCount/deltaTimeInSeconds,
			totalEntry.loggedCount-totalEntry.collectedCount,
		)

	}
	fmt.Printf("\n\n")
}
