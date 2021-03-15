package main

import (
    "encoding/json"
    "errors"
    "flag"
    "fmt"
    "github.com/papertrail/go-tail/follower"
    "io"
    "log"
    "sort"
    "strconv"
    "strings"
    "time"
)

type logSourceInfo struct  {
    loggedCount             int64
    collectedCount          int64

    firstSeq                int64
}

type reportStatistics struct  {
    totalLogsCollectedCount         int64
    startMonitoringTime                 time.Time
}

func main() {

    var fluentLogFileName string
    var reportCount int64

    flag.StringVar(&fluentLogFileName, "f", "0.log", "fluent log file to tail")
    flag.Int64Var(&reportCount, "c", 100, "number of logs between reports")
    flag.Parse()

    logsCurrentInfo := make(map[string]logSourceInfo)
    logsTotalInfo := make(map[string]logSourceInfo)
    reportData := reportStatistics{}

    reportData.startMonitoringTime = time.Now()

    // forever loop tailing on the file
    for {
        // log.Printf("Watching file: %s",fluentLogFileName)
        time.Sleep(1 * time.Millisecond)

        t, err := follower.New(fluentLogFileName , follower.Config{
            Whence: io.SeekEnd,
            Offset: 0,
            Reopen: true,
        })

        if err != nil {
            // log.Printf("follower.New: Error  %v",err)
            continue
        }

        for line := range t.Lines() {
            errParse, name, seq := parseLine(line.String())
            if errParse != nil {
                // log.Printf("Error in Line: %s", line.String())
                // log.Printf("%v",errParse)
                continue
            }

            if _, ok := logsCurrentInfo[name]; !ok {
                logsCurrentInfo[name] = logSourceInfo{}
            }

            if _, ok := logsTotalInfo[name]; !ok {
                logsTotalInfo[name] = logSourceInfo{}
            }

            // calculate current metrics
            // TODO: Make sure we recognize concat `logtag` values (e.g. 'P')
            // TODO: and handle partial lines == do not count split log lines twice
            entry:= logsCurrentInfo[name]
            if entry.firstSeq == 0 {
                entry.firstSeq = seq-1
            }
            entry.collectedCount += 1
            entry.loggedCount = seq - entry.firstSeq

            // calculate total metrics
            totalEntry:= logsTotalInfo[name]
            if totalEntry.firstSeq == 0 {
                totalEntry.firstSeq = seq-1
            }
            totalEntry.collectedCount += 1
            totalEntry.loggedCount = seq - totalEntry.firstSeq

            // calculate global metrics
            reportData.totalLogsCollectedCount +=1

            // persist
            logsCurrentInfo[name] = entry
            logsTotalInfo[name] = totalEntry

            if reportData.totalLogsCollectedCount % reportCount == 0 {
                report(reportData, logsCurrentInfo, logsTotalInfo)

                // reset counting
                logsCurrentInfo = make(map[string]logSourceInfo)
            }
        }
    }
}
func parseLine(line string) (err error,name string, seq int64) {
    lineSplit := strings.Split(line, "{")
    if len(lineSplit) < 2 {
        err = errors.New("parseLine: split { failed")
        return err, "", 0
    }
    jsonString := "{"+strings.Split(lineSplit[1],"}")[0]+"}"
    var j map[string]interface{}
    if err1 := json.Unmarshal([]byte(jsonString), &j); err1 != nil {
        err = errors.New("parseLine: split } failed")
        return err, "", 0
    }

    if _, ok := j["path"]; !ok {
        err = errors.New("parseLine: can't find path")
        return err, "", 0
    }

    if _, ok := j["log"]; !ok {
        err = errors.New("parseLine: can't find log")
        return err, "", 0
    }

    // get the file name of log (container name)
    nameSliced := strings.Split(fmt.Sprintf("%s", j["path"]), "_")
    if len(nameSliced) < 1 {
        err = errors.New("parseLine: can't parse _ in path")
        return err, "", 0
    }
    nameSliced = strings.Split(fmt.Sprintf("%s", nameSliced[0]), "/")
    if len(nameSliced) < 5 {
        err = errors.New("parseLine: can't parse / in path")
        return err, "", 0
    }

    if nameSliced[3] != "containers" {
        err = errors.New("parseLine: can't parse / path -> follow only  /var/log/containers ")
        return err, "", 0
    }

    name = nameSliced[4]

    // get the sequence number of the log
    logSliced := strings.Split(fmt.Sprintf("%s", j["log"]), "-")
    if len(logSliced) < 3 {
        err = errors.New("parseLine: can't parse - in log")
        return err, "", 0
    }
    seqStr := strings.TrimSpace(logSliced[2])
    seq, err1 := strconv.ParseInt(seqStr, 10, 0)
    if err1 != nil{
        err = errors.New("parseLine: can't parse ParseInt in log")
        return err, "", 0
    }

    return nil, name, seq
}

func report(reportData reportStatistics, logsCurrentInfo map[string]logSourceInfo, logsTotalInfo map[string]logSourceInfo ) {
    now := time.Now()
    deltaTimeInSeconds := now.Unix() - reportData.startMonitoringTime.Unix()
    log.Printf("Report at: %s\n", now.String())
    log.Printf("-==-=-=-=-=\n")
    if  !reportData.startMonitoringTime.IsZero() {
        log.Printf("Total number of collected logs : %d\n", reportData.totalLogsCollectedCount)
        log.Printf("Logs per sec : %d\n", reportData.totalLogsCollectedCount / deltaTimeInSeconds)
        log.Printf("Time from start monitoring (in secs): %d\n", deltaTimeInSeconds)
    }

    log.Printf("-==-=-=-=-=\n")
    tableFormat := "| %-36v | %-9v | %-9v | %-9v | %-9v | %-9v | %-9v | %-9v | %-9v |\n"
    tableFormatLen := len(fmt.Sprintf(tableFormat,0,0,0,0,0,0,0,0,0))-1
    log.Printf(strings.Repeat("-", tableFormatLen))
    log.Printf(tableFormat,
        "",
        "Current",
        "Lines",
        "",
        "Total",
        "Lines",
        "",
        "",
        "")
    log.Printf(strings.Repeat("-", tableFormatLen))
    log.Printf(tableFormat,
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
    log.Printf(strings.Repeat("-", tableFormatLen))

    names := make([]string, 0)
    for k := range logsTotalInfo {
        names = append(names, k)
    }
    sort.Strings(names)

    for _, name := range names {
        entry := logsCurrentInfo[name]
        totalEntry := logsTotalInfo[name]
        log.Printf(tableFormat,name,
            entry.loggedCount,
            entry.collectedCount,
            entry.loggedCount - entry.collectedCount,
            totalEntry.loggedCount,
            totalEntry.loggedCount / deltaTimeInSeconds,
            totalEntry.collectedCount,
            totalEntry.collectedCount / deltaTimeInSeconds,
            totalEntry.loggedCount - totalEntry.collectedCount,
            )

    }
    log.Printf("\n\n")
}
