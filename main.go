package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	metrics         = "metrics"
	info            = "info"
	logsDir         = "/var/log/mc_users_bd"
	pollingInterval = 5000 * time.Millisecond
)

// Metric type defines the info to be written in a metric log trace
type Metric struct {
	// Level is the log level
	Level string `json:"lvl"`
	// MetricTime is the timestamp at the log has been written
	MetricTime string `json:"time"`
	// Msg is the log message
	Msg string `json:"msg"`
	// Full name of collection of which we are counting its documents
	CollectionName string `json:"collection_name"`
	// The number of documents in CollectionName
	CountValue int `json:"count_value"`
}

var (
	logger *log.Logger
)

func main() {
	//session, err := mgo.Dial("server1.example.com, server2.example.com")
	config, err := getConfiguration("config.json")
	if err != nil {
		panic(err)
	}

	createDirIfNotExist(logsDir)
	logFile, err := os.OpenFile(filepath.Join(logsDir, "mc_users_bd.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	/*****************************************************/
	counterChan := make(chan Value)
	for _, counter := range config.Counters {
		go func() {
			counter.getCounters(counterChan)
		}()
	}

	logger = log.New(logFile, info, log.Ldate|log.Ltime|log.Lshortfile)
	for {
		//fmt.Println("Lo que viene del canal es: ", <-counterChan)
		counterValue := <-counterChan
		metric := Metric{MetricTime: time.Now().Format(time.RFC3339), Level: info, Msg: metrics, CollectionName: counterValue.name, CountValue: counterValue.value}
		jsonMetric, _ := json.Marshal(metric)
		//os.Stdout.Write(jsonMetric)
		fmt.Println(string(jsonMetric))
		logger.Println(string(jsonMetric))
	}

}

func createDirIfNotExist(dirName string) {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		if err := os.Mkdir(dirName, 0666); err != nil {
			panic(err)
		}
	}
}
