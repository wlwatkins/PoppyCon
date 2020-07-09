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

func CheckErr(err error, errMsg string) {
  if err != nil {
    log.Fatal(fmt.Sprintf(errMsg, err))
  }
}


func queryDB(p string) *sql.Rows {
  db, err := sql.Open("sqlite3", "./data.db")
  CheckErr(err, "Could not open database: %s")
  r, err := db.Query(fmt.Sprintf("SELECT * FROM sensors WHERE sensorID = '%s' ORDER BY date ASC", p))
  CheckErr(err, "Error whilst doing query: %s")
  db.Close()
  return r
}

type CalibCSV struct {
    ZeroPCT string
    HundredPCT string
}

func getCalib() map[string]CalibCSV{
  f, err := os.Open("utils/calibration.config")
  CheckErr(err, "Could not read calibration.config: %s")
  r := csv.NewReader(bufio.NewReader(f))

  op := make(map[string]CalibCSV)
  for {
        l, err := r.Read()
        if err == io.EOF {
            break
        } else if err != nil {
          CheckErr(err, "Error whilst reading file: %s")
        }
        op[l[1]] =  CalibCSV{
            ZeroPCT: l[2],
            HundredPCT: l[3],
        }
    }
    return op
}

func sensorList() []string {
  f, err := os.Open("utils/sensors.config")
  CheckErr(err, "Could not read sensors.config: %s")
  r := csv.NewReader(bufio.NewReader(f))

  var op []string
  for {
        l, err := r.Read()
        if err == io.EOF {
            break
        } else if err != nil {
          CheckErr(err, "Error whilst reading file: %s")
        }
        op = append(op, l[0])
    }
    return op
}

type moisterProbe struct {
  Date int64
  Voltage float64
  Value int
}

type MasterStruct struct {
  Calib CalibCSV
  Data []moisterProbe
}

func readMoister() map[string][]map[string]MasterStruct {
    calib := getCalib()
    probesList := sensorList()

    // database row decleration
    var id int
    var sensorType string
    var sensorID string
    var date int64
    var valueInt int
    var valueFloat float64
    var name string
    var desciption string

    var dataProbeSlice []moisterProbe
    sensors := make(map[string][]map[string]MasterStruct)

    for _, probeID := range probesList {
      probeMaster := make(map[string]MasterStruct)
      rows := queryDB(probeID)
      dataProbeSlice = make([]moisterProbe, 0)
      for rows.Next() {
          rows.Scan(&id, &sensorType, &sensorID, &date, &valueFloat, &valueInt, &name, &desciption)
          dataProbeSlice = append(dataProbeSlice, moisterProbe{Date: date, Voltage: valueFloat, Value: valueInt})
      }
      probeMaster[probeID] = MasterStruct{Calib: calib[probeID], Data: dataProbeSlice}
      sensors[sensorType] = append(sensors[sensorType], probeMaster)
    }
    return sensors
  }
