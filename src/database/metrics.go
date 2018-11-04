package database

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
)

// Metric stores metric collected from a function call
type Metric struct {
	function                  string
	containerID               string
	containerCreateTime       float64
	containerStartTime        float64
	applicationConnectionTime float64
	applicationRunTime        float64
	applicationCode           string
	containerStopTime         float64
	containerDeleteTime       float64
}

// MetricDB stores all metrics
type MetricDB struct {
	storePath string
	metrics   []Metric
}

// NewMetricBD creates a metric database that stores data in the specified path
func NewMetricBD(storePath string) MetricDB {
	mdb := MetricDB{storePath: storePath, metrics: make([]Metric, 0)}
	if _, err := os.Stat(storePath); os.IsExist(err) {
		fileReader, _ := os.Open(storePath)
		json.NewDecoder(fileReader).Decode(&mdb.metrics)
	}
	return mdb
}

// StartMetricDBRoutine starts the MetricDB subroutine
func (mdb *MetricDB) StartMetricDBRoutine() (chan Metric, chan bool) {
	metricChannel := make(chan Metric, 1000)
	persistChannel := make(chan bool)

	go mdb.metricDBRoutine(metricChannel, persistChannel)

	return metricChannel, persistChannel
}

func (mdb *MetricDB) metricDBRoutine(metricChannel chan Metric, persistChannel chan bool) {
	for {
		// save metric
		metric := <-metricChannel
		if len(metric.function) != 0 {
			mdb.metrics = append(mdb.metrics, metric)
		}

		select {
		case <-persistChannel:
			mdb.storeMetrics()
		default:
		}
	}
}

func (mdb *MetricDB) storeMetrics() {
	buffer := bytes.Buffer{}
	json.NewEncoder(&buffer).Encode(mdb.metrics)
	ioutil.WriteFile(mdb.storePath, buffer.Bytes(), 644)
}
