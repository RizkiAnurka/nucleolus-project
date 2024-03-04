package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

var every1MinuteLog string

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
	InitializeEvery1MinuteJob()

}

// Entry Point from the App Init
func InitializeEvery1MinuteJob() {

	every1MinuteLog = fmt.Sprint("job_everyminute_", time.Now().Format("20060102"), ".log")
	valid := ValidateLastJobExecuted(every1MinuteLog, 60)

	if valid != true {
		batchStamp := time.Now().UnixNano()
		Every1MinuteJob(batchStamp)
	}
}

// Entry Point from the scheduler
func RunEvery1MinuteJob() {
	executionStart := time.Now().UnixNano()
	Every1MinuteJob(executionStart)
}

// Entry Point from the endpoint
func TriggerEvery1MinuteJob(http.ResponseWriter, *http.Request) {
	executionStart := time.Now().UnixNano() //we could execute payload from endpoint
	Every1MinuteJob(executionStart)
}

// Handler for last job checking
func RetrieveSchedulerLog(file string) string {
	fileHandle, err := os.Open(file)
	if err != nil {
		fileHandle, _ := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		defer fileHandle.Close()
		writeLog(file, "Log file created.")
		return ""
	}
	defer fileHandle.Close()
	line := ""
	var cursor int64 = 0
	stat, _ := fileHandle.Stat()
	filesize := stat.Size()
	for {
		cursor -= 1
		fileHandle.Seek(cursor, io.SeekEnd)

		char := make([]byte, 1)
		fileHandle.Read(char)

		if cursor != -1 && (char[0] == 10 || char[0] == 13) { // stop if we find a line
			break
		}

		line = fmt.Sprintf("%s%s", string(char), line) // there is more efficient way

		if cursor == -filesize { // stop if we are at the begining
			break
		}
	}

	return line
}

// Validation of Log, either the init should rerun the job or not
func ValidateLastJobExecuted(schedulerLog string, expectedDiff int) bool {
	lastLine := RetrieveSchedulerLog(schedulerLog)
	log.Println("Last line of log file: ", lastLine)
	lastJob, err := strconv.ParseInt(lastLine, 10, 64)
	if err != nil {
		_, fn, line, _ := runtime.Caller(1)
		log.WithFields(log.Fields{
			"filename": fn,
			"line":     line,
			"param":    lastLine,
		}).Error(err.Error())
	}

	lastUnix := time.Unix(0, lastJob)
	actualDiff := time.Since(lastUnix).Minutes()
	if int(actualDiff) >= expectedDiff {
		return true
	}
	return false
}

// Write log to file
func writeLog(file string, message string) {
	fileHandle, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Panic("Failed to open log file", err)
		return
	}

	defer fileHandle.Close()

	_, err2 := fileHandle.Write([]byte(fmt.Sprintf("\n%s", message)))

	if err2 != nil {
		log.Error("Could not write to file", err2)

	} else {
		log.Info("Operation successful! Text has been appended")
	}
}

func main() {
	SetupScheduler()

	// Registering our handler functions, and creating paths.

	http.HandleFunc("/job/every1minute/", TriggerEvery1MinuteJob)

	log.Println("To close connection CTRL+C :-)")

	// Spinning up the server.
	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// Setup Scheduler, this function should be call in the Service init
func SetupScheduler() {
	jakartaTime, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		_, fn, line, _ := runtime.Caller(1)
		log.WithFields(log.Fields{
			"filename": fn,
			"line":     line,
			"param":    "Asia/Jakarta",
		}).Error(err.Error())
	}
	scheduler := cron.New(cron.WithLocation(jakartaTime))

	defer scheduler.Stop()

	scheduler.AddFunc("*/1 * * * *", func() {
		RunEvery1MinuteJob()
	})

	// start scheduler
	go scheduler.Start()
}

// The job function
func Every1MinuteJob(executionStart int64) {
	// write log before job is done
	log.Info("Start RunEvery1MinuteJob Scheduler ", executionStart)

	//execute job
	log.Info("Running job every 1 minute")

	// write log after job is done
	writeLog(every1MinuteLog, fmt.Sprintf("%v", executionStart))
	end := time.Now().UnixNano() / int64(time.Millisecond)
	diff := end - (executionStart / int64(time.Millisecond))
	log.Info("Finish RunEvery1MinuteJob Scheduler ", executionStart, " Elapsed Time : ", diff, "ms")
}
