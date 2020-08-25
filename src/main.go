package main

import (
  "gobot.io/x/gobot"
  "gobot.io/x/gobot/drivers/i2c"
  "os"
  "database/sql"
  _ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

var (
  db *sql.DB
  v float64
  pumpChan chan int
  tempsChanIn chan int
  tempsChanOut chan map[string]float64
  doMeaChan chan int
  relay1State bool
  relay2State bool
  hum *i2c.ADS1x15Driver
  light *i2c.ADS1x15Driver
)


func main() {

  _, err := os.Stat("sqlite-database.db")
  db, _ = sql.Open("sqlite3", "sqlite-database.db")
  if os.IsNotExist(err) {
    createTable(db)
  }
  defer db.Close() // Defer Closing the database
  pumpChan = make(chan int)
  master := gobot.NewMaster()


  master.AddRobot(PumpControl())
  master.AddRobot(SensorAcquisition(db))
  master.AddRobot(webServer())

  master.Start()
}
