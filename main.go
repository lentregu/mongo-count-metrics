package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	metrics         = "metrics"
	info            = "info"
	logsDir         = "/var/log/metrics_bd"
	pollingInterval = 5000 * time.Millisecond
)

type DataConfig struct {
	LogsDir         string `json:"logsDir"`
	LogsFile        string `json:"logsFile"`
	PollingInterval int    `json:"pollingInterval"`
	Metrics_msg     string `json:"metrics_msg"`
}

type Config struct {
	fileName   string
	dataConfig DataConfig
}

func (c *Config) read() error {
	// Nota: intentar hacer genérico con interfaz para Config y BDConfig
	configFile, _ := os.Open(c.fileName)
	jsonDecoder := json.NewDecoder(configFile)
	c.dataConfig = DataConfig{}
	err := jsonDecoder.Decode(&c.dataConfig)
	return err
}

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
	logger   *log.Logger
	logFile  *os.File
	bdConfig *BDConfig
	config   *Config
)

func init() {

	var fatalErr error
	defer func() {
		if fatalErr != nil {
			flag.PrintDefaults()
			log.Fatalln(fatalErr)
		}
	}()

	configFileName := flag.String("config-filename", "config.json", "the application config file")
	bdConfigFileName := flag.String("bd-filename", "bd.json", "the configuration of the bd counters")

	//flag.Args --> Return the non-flag command-line arguments
	args := flag.Args()
	flag.Parse()

	if len(args) != 0 {
		fatalErr = errors.New("invalid usage; must specify command")
		return
	}

	var err error

	config = &Config{fileName: *configFileName}
	err = config.read()
	if err != nil {
		fatalErr = errors.New("Error, the file " + *configFileName + " is not present")
		return
	}

	bdConfig = &BDConfig{fileName: *bdConfigFileName}
	err = bdConfig.read()
	if err != nil {
		fatalErr = errors.New("Error, the file " + *bdConfigFileName + " is not present")
		return
	}

	createDirIfNotExist(config.dataConfig.LogsDir)
	logFile, err = os.OpenFile(filepath.Join(config.dataConfig.LogsDir, config.dataConfig.LogsFile), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fatalErr = errors.New(err.Error())
		return
	}

}

func createDirIfNotExist(dirName string) {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		if err := os.Mkdir(dirName, 0666); err != nil {
			panic(err)
		}
	}
}

func main() {

	counterChan := make(chan Value)
	for _, counter := range bdConfig.dataBDConfig.Counters {
		go func() {
			counter.getCounters(counterChan)
		}()
	}

	//logger = log.New(logFile, info, log.Ldate|log.Ltime|log.Lshortfile)
	for {
		counterValue := <-counterChan
		metric := Metric{MetricTime: time.Now().Format(time.RFC3339), Level: info, Msg: config.dataConfig.Metrics_msg, CollectionName: counterValue.name, CountValue: counterValue.value}
		jsonMetric, _ := json.Marshal(metric)
		//os.Stdout.Write(jsonMetric)
		logFile.Write(jsonMetric)
		logFile.WriteString("\n")
	}

}
