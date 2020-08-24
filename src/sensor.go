package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "time"
    "errors"

    "gobot.io/x/gobot"
    "gobot.io/x/gobot/drivers/i2c"
    "gobot.io/x/gobot/platforms/raspi"

    "github.com/MichaelS11/go-dht"

    "database/sql"
    _ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)


type Results struct {
	Results Result `json:"results"`
}

type Result struct {
	Sunrise string `json:"sunrise"`
  Sunset string `json:"sunset"`
}

type temps struct {
  ChannelIn chan int
  ChannelOut chan map[string]float64

}


func readOneWire(t temps){

	names, err := ScanSlaves()
	if err != nil {
		fmt.Printf("Error scanning 1wire names: %v", err)
		return
	}

	devices := make([]*DS18S20, len(names))
	for i := range names {
		devices[i], err = NewDS18S20(names[i])
		if err != nil {
			fmt.Printf("Error opening device %v: %v", devices[i], err)
			return
		}
	}

	for {
    probe := make(map[string]float64)
    select {
    case <- t.ChannelIn:

  		for i := range devices {
  			value, err := devices[i].Read()
  			if err != nil {
  				fmt.Printf("Error on read: %v", err)
  			} else {
          probe[devices[i].Name] = float64(value)/1000
  			}
  			// time.Sleep(time.Duration(10) * time.Millisecond)
  		}
      t.ChannelOut <- probe
    }
	}
}

func getTemps(tempsChan temps) []dbSensorRow {

  data := []dbSensorRow{}

  // Initiating temperature sensors (1Wire)
  tempsChan.ChannelIn <- 1
  temps := <- tempsChan.ChannelOut

  for id, v := range temps{
    data = append(data, dbSensorRow{ sensorType: "TEMPERATURE" ,
                  sensorID: id,
                  date: time.Now().Unix(),
                  valueFloat: v,
                })
  }

  return data
}

func getI2CSensor(sensorI2C *i2c.ADS1x15Driver, valType int) []dbSensorRow {

  var sensorType string
  data := []dbSensorRow{}

  switch valType {
    case 1:
      sensorType = "MOISTURE"
    case 2:
      sensorType = "LIGHT"
  }

  for i := 0; i<4; i++{
    v, _ = sensorI2C.ReadWithDefaults(i)
    data = append(data, dbSensorRow{ sensorType: sensorType,
                  sensorID: fmt.Sprintf("prob%v%v", valType, i),
                  date: time.Now().Unix(),
                  valueFloat: v,
                })
  }

  return data
}


func readDHT22(dht *dht.DHT) ([]dbSensorRow, error) {
  data := []dbSensorRow{}
  humidity, temperature, err := dht.Read()
  if err != nil {
    fmt.Println("Read error:", err)
    return data, errors.New("can't read DHT")
  }

  data = append(data, dbSensorRow{ sensorType: "dhtHum",
                sensorID: "dhtHum",
                date: time.Now().Unix(),
                valueFloat: humidity,
              })

  time.Sleep(1 * time.Second)

  data = append(data, dbSensorRow{ sensorType: "dhtTemp",
                sensorID: "dhtTemp",
                date: time.Now().Unix(),
                valueFloat: temperature,
              })
  return data, nil
}



func readDayNight(when time.Time) dbDayTimeRow {
  lat := 48.877063
  lng := 2.293546
  loc, _ := time.LoadLocation("Europe/Paris")
  now := time.Now()
  today := when.Format("2006-01-02")

  var results Results
  url := fmt.Sprintf("https://api.sunrise-sunset.org/json?lat=%v&lng=%v&date=%s", lat, lng, today)
  fmt.Println(url)
  apiClient := http.Client{
		Timeout: time.Second * 10, // Timeout after 2 seconds
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

  return dbDayTimeRow{CallTime: now.Unix(),
                  Date: when.Truncate(24*time.Hour).Unix(),
                  SunRiseEpoch: sunRiseEpoch,
                  SunSetEpoch: sunSetEpoch,
                  DayTimeEpoch: sunSetEpoch-sunRiseEpoch}

}

func updateDayNight() {

  var firstDay time.Time
  var dtEpochs dbDayTimeRow
  i := 0
  dtEpochs = readDayNight(time.Now())
  insertDayTimeDB(dtEpochs)

  datesObj := readDayNightDB()
  firstDay = time.Unix(datesObj[i].Date, 0)

  for i < len(datesObj) { //_, dObj := range datesObj {
    y, m, d := time.Unix(datesObj[i].Date, 0).Date()
    yRef, mRef, dRef := firstDay.Date()

    if y != yRef || m != mRef || d != dRef {
      dtEpochs = readDayNight(firstDay)
      insertDayTimeDB(dtEpochs)
    } else {
      i++
    }
    firstDay = firstDay.AddDate(0, 0, 1)
  }
}

func calibSensor(sensor string) float64{

  var data []dbSensorRow
  v := 0.0
  switch sensor{
  case "prob10":
    data = getI2CSensor(hum, 1)
    return data[0].valueFloat
  case "prob11":
    data = getI2CSensor(hum, 1)
    return data[1].valueFloat
  case "prob12":
    data = getI2CSensor(hum, 1)
    return data[2].valueFloat
  case "prob13":
    data = getI2CSensor(hum, 1)
    return data[3].valueFloat
  case "prob20":
    data = getI2CSensor(light, 2)
    return data[0].valueFloat
  case "prob21":
    data = getI2CSensor(light, 2)
    return data[0].valueFloat
  case "prob22":
    data = getI2CSensor(light, 2)
    return data[0].valueFloat
  case "prob23":
    data = getI2CSensor(light, 2)
    return data[0].valueFloat
  }

  return v
}

func SensorAcquisition(db *sql.DB) *gobot.Robot {
  rpi := raspi.NewAdaptor()

  var data []dbSensorRow


  tempsChanIn = make(chan int)
  tempsChanOut = make(chan map[string]float64)
  tempsChan := temps{ ChannelIn: tempsChanIn,
                      ChannelOut: tempsChanOut}

  // Initiation moisture sensors (I2C)
  hum = i2c.NewADS1115Driver(rpi, i2c.WithAddress(0x48))
	hum.DefaultGain, _ = hum.BestGainForVoltage(5.0)

  // Initiation light sensors (I2C)
  light = i2c.NewADS1115Driver(rpi, i2c.WithAddress(0x49))
	light.DefaultGain, _ = light.BestGainForVoltage(5.0)

  // Initilise DHT22 onewire sensor
  err := dht.HostInit()
  if err != nil {
    fmt.Println("HostInit error:", err)
  }

  dht, err := dht.NewDHT("GPIO26", dht.Celsius, "")
  if err != nil {
    fmt.Println("NewDHT error:", err)
  }



  work := func() {

          go readOneWire(tempsChan)
          updateDayNight()
          gobot.Every(900*time.Second, func() {

                  //Get temperatures
                  data = getTemps(tempsChan)
                  for _, d := range data{
                    insertSensorDB(d)
                  }

                  //Get humidity
                  data = getI2CSensor(hum, 1)
                  for _, d := range data{
                    insertSensorDB(d)
                  }

                  //Get light
                  data = getI2CSensor(light, 2)
                  for _, d := range data{
                    insertSensorDB(d)
                  }

                  //Get DHT
                  data, err = readDHT22(dht)
                  for _, d := range data{
                    insertSensorDB(d)
                  }


          })

          gobot.Every(12*time.Hour, func() {

                  updateDayNight()


          })
  }

  robot := gobot.NewRobot("sensors",
          []gobot.Connection{rpi},
          []gobot.Device{hum, light},
          work,
  )

  return robot

}
