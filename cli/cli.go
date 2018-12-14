package main

import (
	"flag"
	"log"
	"strings"

	"github.com/theverything/metrolinkstatus/cmd/metrolink"
)

type stationsFlag []string

func (sf *stationsFlag) String() string {
	return strings.Join(*sf, ", ")
}

func (sf *stationsFlag) Set(value string) error {
	*sf = append(*sf, value)
	return nil
}

func main() {
	var slackWebhookURL = flag.String("slack-webhook", "", "The URL of the slack webhook")
	var debug = flag.Bool("debug", false, "Print debug info.")
	var stations stationsFlag

	flag.Var(&stations, "station", "Station to check times on.")

	flag.Parse()

	stationScheduleList, err := metrolink.LoadStationScheduleList()
	if err != nil {
		log.Fatal(err)
	}

	for _, station := range stations {
		body, err := metrolink.ProcessStation(strings.ToUpper(station), stationScheduleList)
		if err != nil {
			log.Fatal(err)
		}

		err = metrolink.PushTrainStatusToSlack(body, *slackWebhookURL, *debug)
		if err != nil {
			log.Fatal(err)
		}
	}
}
