package main

import (
  "encoding/json"
  	"fmt"
  	"io/ioutil"
  	"log"
  	"net/http"
  	"time"
)

type Results struct {
	Results Result `json:"results"`
}

type Result struct {
	Sunrise string `json:"sunrise"`
  Sunset string `json:"sunset"`
}


func readDayNight() {
  lat := 48.877063
  lng := 2.293546
  loc, _ := time.LoadLocation("Europe/Paris")
  now := time.Now()
  today := now.Format("2006-01-02")

  var results Results
  url := fmt.Sprintf("https://api.sunrise-sunset.org/json?lat=%v&lng=%v&date=%s", lat, lng, today)
  fmt.Println(url)
  apiClient := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

  res, getErr := apiClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

  err = json.Unmarshal(body, &results)
	if err != nil {
		fmt.Println(err)

	}

  sunRiseUTC, _ := time.Parse("2006-01-02 3:04:05 PM", fmt.Sprintf("%s %s", today, results.Results.Sunrise))
  sunSetUTC, _ := time.Parse("2006-01-02 3:04:05 PM", fmt.Sprintf("%s %s", today, results.Results.Sunset))
  sunRise := sunRiseUTC.In(loc)
  sunSet := sunSetUTC.In(loc)
  sunRiseEpoch := sunRise.Unix()
  sunSetEpoch := sunSet.Unix()
  _=sunRiseEpoch
  _ =sunSetEpoch


  fmt.Println(time.Now().Truncate(24*time.Hour).Unix())

}


func main() {
   readDayNight()
}
