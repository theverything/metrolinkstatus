package metrolink

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// ScheduledStop -
type ScheduledStop struct {
	TrainDesignation               string `json:"TrainDesignation"`
	RouteCode                      string `json:"RouteCode"`
	TrainDestination               string `json:"TrainDestination"`
	PlatformName                   string `json:"PlatformName"`
	EventType                      string `json:"EventType"`
	TrainMovementTime              string `json:"TrainMovementTime"`
	CalcTrainMovementTime          string `json:"CalcTrainMovementTime"`
	FormattedTrainMovementTime     string `json:"FormattedTrainMovementTime"`
	FormattedCalcTrainMovementTime string `json:"FormattedCalcTrainMovementTime"`
	FormattedTrackDesignation      string `json:"FormattedTrackDesignation"`
	CalculatedStatus               string `json:"CalculatedStatus"`
	PTCStatus                      string `json:"PTCStatus"`
}

type trainStatusMsg struct {
	Text     string   `json:"text,omitempty"`
	Color    string   `json:"color,omitempty"`
	MrkdwnIn []string `json:"mrkdwn_in,omitempty"`
}

type slackMsg struct {
	Text        string           `json:"text,omitempty"`
	Attachments []trainStatusMsg `json:"attachments,omitempty"`
}

const (
	metrolinkStationStatusURL = "https://rtt.metrolinktrains.com/CIS/LiveTrainMap/JSON/StationScheduleList.json"
)

var trainStatus = map[string]string{
	"ON TIME":          "good",
	"DELAYED":          "warning",
	"EXTENDED DELAYED": "danger",
	"CANCELLED":        "danger",
}

var lineShortName = map[string]string{
	"VC LINE":    "VT",
	"91/PV Line": "91",
	"91PV Line":  "91",
	"AV LINE":    "AV",
	"IE LINE":    "IE",
	"IEOC LINE":  "IE",
	"OC LINE":    "OC",
	"SB LINE":    "SB",
	"VT LINE":    "VT",
}

// MetrolinkStations -
var MetrolinkStations = map[string]string{
	"ANAHEIM-CANYON":            "Anaheim Canyon",
	"ARTIC":                     "Anaheim",
	"BALDWINPARK":               "Baldwin Park",
	"BUENAPARK":                 "Buena Park",
	"BURBANK-AIRPORT-NORTH":     "Burbank Airport - North",
	"BURBANK-AIRPORT-SOUTH":     "Burbank Airport - South",
	"CALSTATE":                  "Cal State LA",
	"CAMARILLO":                 "Camarillo",
	"CHATSWORTH":                "Chatsworth",
	"CLAREMONT":                 "Claremont",
	"COMMERCE":                  "Commerce",
	"COVINA":                    "Covina",
	"DOWNTOWN BURBANK":          "Burbank - Downtown",
	"ELMONTE":                   "El Monte",
	"FONTANA":                   "Fontana",
	"FULLERTON":                 "Fullerton",
	"GLENDALE":                  "Glendale",
	"INDUSTRY":                  "Industry",
	"IRVINE":                    "Irvine",
	"LAGUNANIGUEL-MISSIONVIEJO": "Laguna Niguel/Mission Viejo",
	"LANCASTER":                 "Lancaster",
	"LAUS":                      "L.A. Union Station",
	"MAIN-CORONA-NORTH":         "Corona - North Main",
	"MONTCLAIR":                 "Montclair",
	"MONTEBELLO":                "Montebello/Commerce",
	"MOORPARK":                  "Moorpark",
	"MORENO-VALLEY-MARCH-FIELD": "Moreno Valley/March Field",
	"NEWHALL":                   "Newhall",
	"NORTHRIDGE":                "Northridge",
	"NORWALK/SANTA FE SPRINGS":  "Norwalk/Santa Fe Springs",
	"NORWALK-SANTAFESPRINGS":    "Norwalk/Santa Fe Springs",
	"OCEANSIDE":                 "Oceanside",
	"ONTARIO-EAST":              "Ontario - East",
	"ORANGE":                    "Orange",
	"OXNARD":                    "Oxnard",
	"PALMDALE":                  "Palmdale",
	"PEDLEY":                    "Jurupa Valley/Pedley",
	"PERRIS-DOWNTOWN":           "Perris - Downtown",
	"PERRIS-SOUTH":              "Perris - South",
	"POMONA-DOWNTOWN":           "Pomona - Downtown",
	"POMONA-NORTH":              "Pomona - North",
	"RANCHO CUCAMONGA":          "Rancho Cucamonga",
	"RIALTO":                    "Rialto",
	"RIVERSIDE-DOWNTOWN":        "Riverside - Downtown",
	"RIVERSIDE-HUNTERPARK":      "Riverside - Hunter Park/UCR",
	"RIVERSIDE-LA SIERRA":       "Riverside - La Sierra",
	"SAN BERNARDINO":            "San Bernardino",
	"SANBERNARDINOTRAN":         "San Bernardino-Downtown",
	"SAN CLEMENTE":              "San Clemente",
	"SAN CLEMENTE PIER":         "San Clemente Pier",
	"SAN JUAN CAPISTRANO":       "San Juan Capistrano",
	"SANTA ANA":                 "Santa Ana",
	"SANTA CLARITA":             "Santa Clarita",
	"SIMIVALLEY":                "Simi Valley",
	"SUN VALLEY":                "Sun Valley",
	"SYLMAR/SAN FERNANDO":       "Sylmar/San Fernando",
	"TUSTIN":                    "Tustin",
	"UPLAND":                    "Upland",
	"VAN NUYS":                  "Van Nuys",
	"VENTURA-EAST":              "Ventura - East",
	"VIA PRINCESSA":             "Via Princessa",
	"VINCENT GRADE/ACTON":       "Vincent Grade/Acton",
	"WEST CORONA":               "Corona - West",
	"SAN BERNARDINO-DOWNTOWN":   "San Bernardino - Downtown",
}

func formatArrivalTime(stop ScheduledStop) string {
	timefmt := "3:04 PM"

	scheduledTime := stop.FormattedTrainMovementTime
	scheduledDateTime, err := time.Parse(timefmt, scheduledTime)
	if err != nil {
		return scheduledTime
	}

	expectedTime := stop.FormattedCalcTrainMovementTime
	expectedDateTime, err := time.Parse(timefmt, expectedTime)
	if err != nil {
		return scheduledTime
	}

	scheduleDiff := expectedDateTime.Sub(scheduledDateTime).Minutes()

	if scheduleDiff != 0 {
		return fmt.Sprintf("*%s (%v)*", expectedTime, scheduleDiff)
	}

	return scheduledTime
}

// PushTrainStatusToSlack -
func PushTrainStatusToSlack(body []byte, slackWebhookURL string, debug ...bool) error {
	if len(debug) > 0 && debug[0] {
		log.Println(string(body))
	}

	req, err := http.NewRequest("POST", slackWebhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		bodyErr, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("StatusCode: %v, Body: %s", res.StatusCode, string(bodyErr))
	}

	if len(debug) > 0 {
		resBody, _ := ioutil.ReadAll(res.Body)
		log.Println(string(resBody))
	}

	return nil
}

// ProcessStation -
func ProcessStation(station string, stationScheduleList []ScheduledStop, filter ...bool) ([]byte, error) {
	var trainStatusMsgs []trainStatusMsg
	shouldFilter := true

	if len(filter) > 0 {
		shouldFilter = filter[0]
	}

	for _, stop := range stationScheduleList {
		calcStatus := trainStatus[stop.CalculatedStatus]
		include := !shouldFilter || calcStatus != "good"

		if stop.PlatformName == station && include {
			trainStatusMsgs = append(trainStatusMsgs, trainStatusMsg{
				Text: fmt.Sprintf("%-17s %-2s %-6s %s on %s",
					formatArrivalTime(stop),
					lineShortName[stop.RouteCode],
					stop.TrainDesignation,
					stop.TrainDestination,
					stop.FormattedTrackDesignation,
				),
				Color:    calcStatus,
				MrkdwnIn: []string{"text"},
			})
		}
	}

	if len(trainStatusMsgs) == 0 {
		trainStatusMsgs = append(trainStatusMsgs, trainStatusMsg{
			Text: "Trains are on time.",
		})
	}

	message := slackMsg{
		Text:        fmt.Sprintf("%s Station - Scheduled Trains", MetrolinkStations[station]),
		Attachments: trainStatusMsgs,
	}

	body, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	return body, err
}

// LoadStationScheduleList -
func LoadStationScheduleList() ([]ScheduledStop, error) {
	var stationScheduleList []ScheduledStop

	metroResp, err := http.Get(metrolinkStationStatusURL)
	if err != nil {
		return nil, err
	}

	defer metroResp.Body.Close()

	if err = json.NewDecoder(metroResp.Body).Decode(&stationScheduleList); err != nil {
		return nil, err
	}

	return stationScheduleList, nil
}
