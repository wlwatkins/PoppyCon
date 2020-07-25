package main

import (
    "time"
    "fmt"
    "gobot.io/x/gobot"
    "gobot.io/x/gobot/drivers/i2c"
    "gobot.io/x/gobot/platforms/raspi"

    "database/sql"
    _ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)


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



func SensorAcquisition(db *sql.DB) *gobot.Robot {
  rpi := raspi.NewAdaptor()
  var data dbSensorRow

  // Initiating temperature sensors (1Wire)
  tempsChanIn = make(chan int)
  tempsChanOut = make(chan map[string]float64)
  tempsChan := temps{ ChannelIn: tempsChanIn,
                      ChannelOut: tempsChanOut}

  // Initiation moisture sensors (I2C)
  hum := i2c.NewADS1115Driver(rpi, i2c.WithAddress(0x48))
	hum.DefaultGain, _ = hum.BestGainForVoltage(5.0)

  // Initiation light sensors (I2C)
  light := i2c.NewADS1115Driver(rpi, i2c.WithAddress(0x49))
	light.DefaultGain, _ = light.BestGainForVoltage(5.0)

  work := func() {

          go readOneWire(tempsChan)

          gobot.Every(900*time.Second, func() {
                  // pumpChan <- 1
                  // pumpChan <- 2
                  tempsChan.ChannelIn <- 1
                  temps := <- tempsChan.ChannelOut

                  //Get temperatures
                  for id, v := range temps{
                    data = dbSensorRow{ sensorType: "TEMPERATURE" ,
                                  sensorID: id,
                                  date: time.Now().Unix(),
                                  valueFloat: v,
                                }
                    insertSensorDB(data)

                  }

                  //Get humidity and light
                  for i := 0; i<4; i++{
                    v, _ = hum.ReadWithDefaults(i)
                    data = dbSensorRow{ sensorType: "MOISTURE" ,
                                  sensorID: fmt.Sprintf("prob1%v", i),
                                  date: time.Now().Unix(),
                                  valueFloat: v,
                                }
                    insertSensorDB(data)
                    v, _ = light.ReadWithDefaults(i)
                    data = dbSensorRow{ sensorType: "LIGHT" ,
                                  sensorID: fmt.Sprintf("prob2%v", i),
                                  date: time.Now().Unix(),
                                  valueFloat: v,
                                }
                    insertSensorDB(data)
                  }
          })
  }

  robot := gobot.NewRobot("sensors",
          []gobot.Connection{rpi},
          []gobot.Device{hum, light},
          work,
  )

  return robot

}
