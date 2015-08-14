package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/mgo.v2"
)

type DataBase struct {
	Name        string
	Collections []string
}

type Counter struct {
	Server string
	Port   int
	Dbs    []DataBase
}

// Config represent all the collections we want to count its documents (bd server, port and an array with the collections)
type DataBDConfig struct {
	Counters []Counter
}

type BDConfig struct {
	fileName     string
	dataBDConfig DataBDConfig
}

type Value struct {
	name  string
	value int
}

func getConfiguration(fileName string) (BDConfig, error) {
	configFile, _ := os.Open(fileName)
	jsonDecoder := json.NewDecoder(configFile)
	config := BDConfig{}
	err := jsonDecoder.Decode(&config)
	return config, err
}

func (c *BDConfig) read() error {
	// Nota: intentar hacer genérico con interfaz para Config y BDConfig
	configFile, _ := os.Open(c.fileName)
	jsonDecoder := json.NewDecoder(configFile)
	c.dataBDConfig = DataBDConfig{}
	fmt.Printf("Filename: %s", c.fileName)
	err := jsonDecoder.Decode(&c.dataBDConfig)
	return err
}

func (c *Counter) getCounters(counterChan chan Value) {
	mongoSession, err := mgo.Dial(c.Server + ":" + strconv.Itoa(c.Port))
	if err != nil {
		panic(err)
	}
	defer mongoSession.Close()
	//Optional. Switch the session to a monotonic behavior.
	mongoSession.SetMode(mgo.Monotonic, true)

	//collectionSessions := make([]*mgo.Collection, 5)
	var collectionSessions []*mgo.Collection
	for _, db := range c.Dbs {
		for _, coll := range db.Collections {
			collectionSessions = append(collectionSessions, mongoSession.DB(db.Name).C(coll))
		}
	}

	for {

		for _, collSession := range collectionSessions {
			sizeColl, _ := collSession.Count()
			counterChan <- Value{name: collSession.FullName, value: sizeColl}
		}

		time.Sleep(time.Duration(config.dataConfig.PollingInterval) * time.Millisecond)
	}

	//close(counterChan)
}
