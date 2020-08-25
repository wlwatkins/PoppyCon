package main

import (
    "database/sql"
    "fmt"
    "os"
    "bufio"
    "encoding/csv"
    "io"
    "log"

    _ "github.com/mattn/go-sqlite3"
)


type dbSensorRow struct {
  sensorType string
  sensorID string
  date int64
  valueFloat float64
}

type dbDayTimeRow struct {
  CallTime int64
  Date int64
  SunRiseEpoch int64
  SunSetEpoch int64
  DayTimeEpoch int64
}

type dbWaterRow struct {
  which int
  date int64
}

type CalibCSV struct {
    ZeroPCT string
    HundredPCT string
}

type dataProbe struct {
  Date int64
  Value float64
}



type MasterStruct struct {
  Calib CalibCSV
  Data []dataProbe
}

func CheckErr(err error, errMsg string) {
  if err != nil {
    log.Print(fmt.Sprintf(errMsg, err))
  }
}




func initDB() *sql.DB {
	sqliteDatabase, _ := sql.Open("sqlite3", "../sqlite-database.db") // Open the created SQLite File
  return sqliteDatabase
}


func createTable(db *sql.DB) {
  var createTableSQL string
  var statement *sql.Stmt
  var err error
  doMeaChan = make(chan int)


	createTableSQL = `CREATE TABLE sensors (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"sensorType" VARCHAR(200),
		"sensorID" VARCHAR(200),
		"date" INTEGER,
    "valueFloat" DECIMAL(10,5)
	  );` // SQL Statement for Create Table

	log.Println("Create table...")
	statement, err = db.Prepare(createTableSQL) // Prepare SQL Statement
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec() // Execute SQL Statements


  createTableSQL = `CREATE TABLE watering (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"date" INTEGER,
		"which" INTEGER
	  );` // SQL Statement for Create Table

	log.Println("Create table...")
	statement, err = db.Prepare(createTableSQL) // Prepare SQL Statement
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec() // Execute SQL Statements

  createTableSQL = `CREATE TABLE dayNight (
    "id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"callTime" INTEGER,
    "date" INTEGER UNIQUE,
		"sunRiseEpoch" INTEGER,
		"sunSetEpoch" INTEGER,
    "dayTimeEpoch" INTEGER
	  );` // SQL Statement for Create Table

	log.Println("Create table...")
	statement, err = db.Prepare(createTableSQL) // Prepare SQL Statement
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec() // Execute SQL Statements

  doMeaChan <- 1 // Send byte to chan to get initial measurements (1: Sensors)
  doMeaChan <- 2 // (2: Day night times)

}

func deleteDB() bool {
  var err = os.Remove("sqlite-database.db")
  return (err != nil)

}

func insertSensorDB(data dbSensorRow) {
	insertSQL := `INSERT INTO sensors(sensorType, sensorID, date, valueFloat) VALUES (?, ?, ?, ?)`
	statement, err := db.Prepare(insertSQL) // Prepare statement.
                                                   // This is good to avoid SQL injections
	if err != nil {
		log.Fatalln(err.Error())
	}

	_, err = statement.Exec(data.sensorType, data.sensorID, data.date, data.valueFloat)
  if err != nil {
		log.Fatalln(err.Error())
	}
}

func insertDayTimeDB(data dbDayTimeRow) {
	insertSQL := `INSERT INTO dayNight(callTime, date, sunRiseEpoch, sunSetEpoch, dayTimeEpoch) VALUES (?, ?, ?, ?, ?)`
	statement, err := db.Prepare(insertSQL) // Prepare statement.
                                                   // This is good to avoid SQL injections
	if err != nil {
		log.Fatalln(err.Error())
	}

	_, err = statement.Exec(data.CallTime, data.Date, data.SunRiseEpoch, data.SunSetEpoch, data.DayTimeEpoch)
  if err != nil {
		log.Print(err.Error())
	}
}

func insertWaterDB(data dbWaterRow) {
	insertSQL := `INSERT INTO watering(date, which) VALUES (?, ?)`
	statement, err := db.Prepare(insertSQL) // Prepare statement.
                                                   // This is good to avoid SQL injections
	if err != nil {
		log.Fatalln(err.Error())
	}

	_, err = statement.Exec(data.date, data.which)
  if err != nil {
		log.Fatalln(err.Error())
	}
}



func getCalib() map[string]CalibCSV{
  var cmt string // variable which will contain first character of each line in file to check if it is commented out
  op := make(map[string]CalibCSV) // map containing calibration data

  //Open calibration file
  f, err := os.Open("utils/calibration.config")
  CheckErr(err, "Could not read calibration.config: %s")

  //Read file in buffer
  r := csv.NewReader(bufio.NewReader(f))

  //Iterate through file
  for {
        l, err := r.Read() //Read line
        if err == io.EOF { //Break if last line
            break
        } else if err != nil {
          CheckErr(err, "Error whilst reading file: %s")
        }
        cmt = l[0][0:1] // Get first character
        if cmt != "#" { // If first character is not # (it is not commented) read the two values
          /*
          l[1]: name of probe
          l[2]: 0% value
          l[3]: 100% value
          */
          op[l[1]] =  CalibCSV{
              ZeroPCT: l[2],
              HundredPCT: l[3],
          }
        }

    }
    return op
}

func sensorList() ([]string, map[string]string) {
  var cmt string // variable which will contain first character of each line in file to check if it is commented out
  var sl []string // raw list of all sensors not commented out
  lm := make(map[string]string) // map containing list of sensors and label

  //Open sensor list file
  f, err := os.Open("utils/sensors.config")
  CheckErr(err, "Could not read sensors.config: %s")
  r := csv.NewReader(bufio.NewReader(f))

  //Iterate through file
  for {
        l, err := r.Read() //Read line
        if err == io.EOF { //Break if last line
            break
        } else if err != nil {
          CheckErr(err, "Error whilst reading file: %s")
        }
        cmt = l[0][0:1] // Get first character
        if cmt != "#" { // If first character is not # (it is not commented) read the two values
          /*
          l[0]: name of probe
          l[1]: label to print on webpage
          */
          lm[l[0]] = l[1]
          sl = append(sl, l[0])
        }
    }
    //returns the raw list of probes and map of probes and labels
    return sl, lm
}

func readSensorDB() map[string][]map[string]MasterStruct {

  //Get calobration data from user file
  calib := getCalib()

  //Get list of probes from user file
  probesList, listMap := sensorList()

  // Variables for columns in database
  var id int
  var sensorType string
  var sensorID string
  var date int64
  var valueFloat float64

  var label string
  var dataProbeSlice []dataProbe
  sensors := make(map[string][]map[string]MasterStruct)

  // Iterate through probe's list
  for _, probeID := range probesList {

    //Get sqlite rows object using SQL query with probeID as variable
    rows, err := db.Query(fmt.Sprintf("SELECT * FROM sensors WHERE sensorID = '%s' ORDER BY date ASC", probeID))
    CheckErr(err, "Error whilst doing query: %s")

    probeMaster := make(map[string]MasterStruct) // Master map that contains the calibration data, raw data and label
    dataProbeSlice = make([]dataProbe, 0) // List of all the data for a given probe. It contains the time in unix epoch and data

    // Iterate each row of the database and append the data to the dataProbeSlice
    for rows.Next() {
        rows.Scan(&id, &sensorType, &sensorID, &date, &valueFloat)
        dataProbeSlice = append(dataProbeSlice, dataProbe{Date: date, Value: valueFloat})
    }

    label = listMap[probeID] // Get label from current probeID
    probeMaster[label] = MasterStruct{Calib: calib[probeID], Data: dataProbeSlice} // Create probeMaster object for given probeID
    sensors[sensorType] = append(sensors[sensorType], probeMaster) // Append probeMaster to sensor slice
  }


  return sensors
}

  type dataWater struct {
    Date int64
    Pump string
  }


func readWaterDB() []dataWater {
    var id int
    var date int64
    var pump string
    var which int64

    var dataWaterSlice []dataWater


    rowsTop, err := db.Query("SELECT * FROM watering ORDER BY date ASC")
    CheckErr(err, "Error whilst doing query: %s")
    for rowsTop.Next() {
        rowsTop.Scan(&id, &date, &which)
        switch which {
        case 1:
          pump = "Top pump"
        case 2:
          pump = "Bottom pump"
        }
        dataWaterSlice = append(dataWaterSlice, dataWater{Date: date, Pump: pump})
    }
    //
    // rowsBottom, err := db.Query("SELECT * FROM watering WHERE which = 2 ORDER BY date ASC")
    // CheckErr(err, "Error whilst doing query: %s")
    // for rowsBottom.Next() {
    //     rowsBottom.Scan(&id, &date, &which)
    //     dataWaterSlice = append(dataWaterSlice, dataWater{Date: date, Value: 2})
    // }

    return dataWaterSlice
  }

func readDayNightDB() []dbDayTimeRow {
    var id int
    var callTime int64
    var date int64
    var sunRiseEpoch int64
    var sunSetEpoch int64
    var dayTimeEpoch int64

    var dataDayNightSlice []dbDayTimeRow


    rowsTop, err := db.Query("SELECT * FROM dayNight ORDER BY date ASC")
    CheckErr(err, "Error whilst doing query: %s")
    for rowsTop.Next() {
        rowsTop.Scan(&id, &callTime, &date, &sunRiseEpoch, &sunSetEpoch, &dayTimeEpoch)
        dataDayNightSlice = append(dataDayNightSlice, dbDayTimeRow{CallTime: callTime, Date: date, SunRiseEpoch: sunRiseEpoch, SunSetEpoch: sunSetEpoch, DayTimeEpoch:dayTimeEpoch})
    }
    return dataDayNightSlice
  }
