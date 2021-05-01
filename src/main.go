package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
)

var (
	DB           *sql.DB
	v            float64
	pumpChan     chan int
	tempsChanIn  chan int
	tempsChanOut chan map[string]float64
	relay1State  bool
	relay2State  bool
	hum          *i2c.ADS1x15Driver
	light        *i2c.ADS1x15Driver
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_, err := os.Stat("sqlite-database.db")
	DB, _ = sql.Open("sqlite3", "sqlite-database.db")
	if os.IsNotExist(err) {
		createTable(DB)
	}
	defer DB.Close() // Defer Closing the database
	pumpChan = make(chan int)
	master := gobot.NewMaster()

	master.AddRobot(PumpControl())
	master.AddRobot(SensorAcquisition(DB))
	master.AddRobot(webServer())

	master.Start()
}
