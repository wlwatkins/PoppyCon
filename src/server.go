package main

import (
  "html/template"
	"io/ioutil"
  "log"
  "net/http"
  "fmt"
  "encoding/json"
  "regexp"
  "time"
  "strings"
  "syscall"
  "gobot.io/x/gobot"
)

type Page struct {
    SensorFile  []byte
    CalibrationFile  []byte
    ParametersFile []byte
    Message []byte
    LastWaterTop []byte
    LastWaterBottom []byte
    Sensor []string
}

var templates = template.Must(template.ParseFiles("tmpl/dashboard.html",
                                                  "tmpl/settings.html",
                                                  "tmpl/calibration.html",
                                                  "tmpl/systemlogs.html",))
var validPath = regexp.MustCompile("^/($|dashboard|settings|pump|calibration|systemlogs|reboot|shutdown)?$")

func CheckErrSend(w http.ResponseWriter,err error, errMsg string) {
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Print(fmt.Sprintf(errMsg, err))
  }
}

func loadFile(file string) ([]byte, error) {
	filename := "utils/" + file
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return []byte(content), nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := templates.ExecuteTemplate(w, tmpl+".html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

type dataAjax struct {
  Data map[string][]map[string]MasterStruct
  Water []dataWater
  DayNight []dbDayTimeRow
  Param string
}

type calibAjax struct {
  Data float64
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "GET":
      // top, bottom := readWaterDB()
      // lastWaterTop := fmt.Sprintf("Last watering of top was : %s", time.Unix(top, 0).Format(time.RFC822Z))
      // lastWaterBottom := fmt.Sprintf("Last watering of bottom was : %s", time.Unix(bottom, 0).Format(time.RFC822Z))
      // p := &Page{LastWaterTop: []byte(lastWaterTop), LastWaterBottom: []byte(lastWaterBottom)}
      p := &Page{}
      renderTemplate(w, "dashboard", p)
    case "POST":
      paramFile, err := loadFile("parameters.config")
      CheckErrSend(w, err, "Could not open parameters: %s")
      jsonData := dataAjax{Data: readSensorDB(), Water: readWaterDB(), DayNight: readDayNightDB(), Param: fmt.Sprintf("%s", paramFile) }
      json.NewEncoder(w).Encode(jsonData)
    default:
      fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
  }
}

func waterHandler(w http.ResponseWriter, r *http.Request) {
  var n int
  switch r.Method {
  case "POST":
    if err := r.ParseForm(); err != nil {
        fmt.Fprintf(w, "ParseForm() err: %v", err)
        return
    }
    whichPump := r.FormValue("button")[0:3]
    switch {
    case whichPump == "top":
      n = 1
    case whichPump == "bot":
      n = 2
    }
    pumpChan <- n
    insertWaterDB(dbWaterRow{date: time.Now().Unix(), which: n})
    // fmt.Println(n)
    gobot.After(10*time.Second, func() {
        // fmt.Println(n)
        pumpChan <- n

    })

    json.NewEncoder(w).Encode("done")
    default:
        fmt.Fprintf(w, "Sorry, only POST methods are supported.")
  }
}

func resetDBHandler(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
  case "POST":
    ok := deleteDB()
    if ok{
      json.NewEncoder(w).Encode("done")
    }else{
      json.NewEncoder(w).Encode("error")
    }

    default:
        fmt.Fprintf(w, "Sorry, only POST methods are supported.")
  }
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "GET":
      calibFile, err := loadFile("calibration.config")
      CheckErrSend(w, err, "Could not open calibration: %s")
      sensorFile, err := loadFile("sensors.config")
      CheckErrSend(w, err, "Could not open sensors: %s")
      paramFile, err := loadFile("parameters.config")
      CheckErrSend(w, err, "Could not open parameters: %s")
      p := &Page{CalibrationFile: calibFile, SensorFile: sensorFile, ParametersFile: paramFile}
      renderTemplate(w, "settings", p)
    case "POST":
        // Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
        if err := r.ParseForm(); err != nil {
            fmt.Fprintf(w, "ParseForm() err: %v", err)
            return
        }
        fileName := r.FormValue("file")
        fileBody := r.FormValue("content")
        err := ioutil.WriteFile("utils/"+fileName, []byte(fileBody), 0600)
        if err != nil {
          msg := fmt.Sprintf("<p><a href='/' > <- Dashboard</a> | Error whilst saving file %s :: %s </p>", fileName, err)
          fmt.Fprintf(w, msg)
        }else{
          msg := fmt.Sprintf("<p>File %s was saved - go back home <a href='/' > <- Dashboard</a></p>", fileName)
          // renderTemplate(w, "success", &Page{Message: []byte(msg)})
          fmt.Fprintf(w, msg)

        }
    default:
        fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
    }

}

func calibrationHandler(w http.ResponseWriter, r *http.Request) {
  var sen []string
  switch r.Method {
    case "GET":
      calibFile, err := loadFile("calibration.config")
      CheckErrSend(w, err, "Could not open calibration: %s")
      sm, _ := sensorList()
      for _, s := range sm{
        if strings.Contains(s, "prob") {
          sen = append(sen, s)
        }
      }
      p := &Page{Sensor: sen, CalibrationFile: calibFile}
      renderTemplate(w, "calibration", p)
    case "POST":
        // Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
        if err := r.ParseForm(); err != nil {
            fmt.Fprintf(w, "ParseForm() err: %v", err)
            return
        }
        sensor := r.FormValue("sensor")
        v := calibSensor(sensor)
        jsonData := calibAjax{Data: v}
        json.NewEncoder(w).Encode(jsonData)

    default:
        fmt.Fprintf(w, "Sorry, only GET methods are supported.")
  }
}

func systemLogsHandler(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "GET":
      logs, err := loadFile("logs")
      CheckErrSend(w, err, "Could not open logs: %s")
      p := &Page{Message: logs}
      renderTemplate(w, "systemlogs", p)
    default:
        fmt.Fprintf(w, "Sorry, only GET methods are supported.")
  }
}

func rebootHandler(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "GET":
      fmt.Fprintf(w, "<p>Rebooting PoppyCon - Wait a few minutes and go back home <a href='/' > <- Dashboard</a></p>")
      gobot.After(5*time.Second, func() {
          // fmt.Println(n)
          err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
          if err != nil{
            log.Println(err)
            fmt.Fprintf(w, "<p>Error, could not reboot <a href='/' > <- Dashboard</a></p>")
          }

      })


    default:
        fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
  }
}

func shutdownHandler(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "GET":
      fmt.Fprintf(w, "<p>Shutting down PoppyCon! You'll need to boot it up manually...</p>")
      gobot.After(5*time.Second, func() {
        err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
        if err != nil{
          log.Println(err)
          fmt.Fprintf(w, "<p>Error, could not shutdown <a href='/' > <- Dashboard</a></p>")
        }
      })


    default:
        fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
  }
}

func makeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r)
	}
}


func webServer() *gobot.Robot  {

  work := func() {


    // go func() {
      http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
      http.HandleFunc("/", makeHandler(dashboardHandler))
      http.HandleFunc("/settings", makeHandler(settingsHandler))
      http.HandleFunc("/calibration", makeHandler(calibrationHandler))
      http.HandleFunc("/systemlogs", makeHandler(systemLogsHandler))
      http.HandleFunc("/reboot", makeHandler(rebootHandler))
      http.HandleFunc("/shutdown", makeHandler(shutdownHandler))
      http.HandleFunc("/pump", makeHandler(waterHandler))


      // http.HandleFunc("/settings/save", makeHandler(settingsSaveHandler))

      log.Fatal(http.ListenAndServe(":80", nil))
       // }
  }



    robot := gobot.NewRobot("webServer",
            work,
    )

    return robot
}
