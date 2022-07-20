package clients

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	EntryApplicationName = "nginx"
	EntrySubsystemName   = "training"
	EntryPrivateKey      = "8a6fe0ef-e3a6-7b4c-30b1-65de28f019c3"
)

type CoralogixClient struct {
}

type message struct {
	PrivateKey      string     `json:"privateKey"`
	ApplicationName string     `json:"applicationName"`
	SubsystemName   string     `json:"subsystemName"`
	LogEntries      []logEntry `json:"logEntries"`
}

type logEntry struct {
	Severity       int     `json:"severity"`
	Timestamp      float64 `json:"timestamp"`
	HiResTimestamp string  `json:"hiResTimestamp"`
	Text           string  `json:"text"`
}

func (c CoralogixClient) SendMsg(text string) {
	url := "https://api.eu2.coralogix.com/api/v1/logs"

	now := time.Now().Unix()
	msg := message{
		PrivateKey:      EntryPrivateKey,
		ApplicationName: EntryApplicationName,
		SubsystemName:   EntrySubsystemName,
		LogEntries: []logEntry{
			{
				Severity:       1,
				Timestamp:      float64(now),
				Text:           text,
				HiResTimestamp: strconv.Itoa(int(now)),
			},
		},
	}

	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(&msg)

	resp, err := http.Post(url, "application/json", payloadBuf)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
}
