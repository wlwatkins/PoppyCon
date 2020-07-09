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
    log.Print(fmt.Sprintf(errMsg, err))
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
  var cmt string
  op := make(map[string]CalibCSV)
  for {
        l, err := r.Read()
        if err == io.EOF {
            break
        } else if err != nil {
          CheckErr(err, "Error whilst reading file: %s")
        }
        cmt = l[0][0:1]
        if cmt != "#" {
          op[l[1]] =  CalibCSV{
              ZeroPCT: l[2],
              HundredPCT: l[3],
          }
        }

    }
    return op
}

func sensorList() ([]string, map[string]string) {
  f, err := os.Open("utils/sensors.config")
  CheckErr(err, "Could not read sensors.config: %s")
  r := csv.NewReader(bufio.NewReader(f))

  lm := make(map[string]string)
  var sl []string
  var cmt string
  for {
        l, err := r.Read()
        if err == io.EOF {
            break
        } else if err != nil {
          CheckErr(err, "Error whilst reading file: %s")
        }
        cmt = l[0][0:1]
        if cmt != "#" {
          lm[l[0]] = l[1]
          sl = append(sl, l[0])
        }
    }
    return sl, lm
}

type moisterProbe struct {
  Date int64
  Voltage float64
  Value int
}

type MasterStruct struct {
  Calib CalibCSV
  Data []moisterProbe
  Label string
}

func readMoister() map[string][]map[string]MasterStruct {
    calib := getCalib()
    probesList, listMap := sensorList()
    fmt.Println(listMap)
    // database row decleration
    var id int
    var sensorType string
    var sensorID string
    var date int64
    var valueInt int
    var valueFloat float64
    var name string
    var desciption string

    var label string
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
      label = listMap[probeID]
      probeMaster[label] = MasterStruct{Calib: calib[probeID], Data: dataProbeSlice}
      sensors[sensorType] = append(sensors[sensorType], probeMaster)
    }
    return sensors
  }
