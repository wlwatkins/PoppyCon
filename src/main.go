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
  "syscall"
)


/*
        syscall commands
        LINUX_REBOOT_CMD_CAD_OFF         = 0x0
        LINUX_REBOOT_CMD_CAD_ON          = 0x89abcdef
        LINUX_REBOOT_CMD_HALT            = 0xcdef0123
        LINUX_REBOOT_CMD_KEXEC           = 0x45584543
        LINUX_REBOOT_CMD_POWER_OFF       = 0x4321fedc
        LINUX_REBOOT_CMD_RESTART         = 0x1234567
        LINUX_REBOOT_CMD_RESTART2        = 0xa1b2c3d4
        LINUX_REBOOT_CMD_SW_SUSPEND      = 0xd000fce2
        LINUX_REBOOT_MAGIC1              = 0xfee1dead
        LINUX_REBOOT_MAGIC2              = 0x28121969
       Linux Man page info
       LINUX_REBOOT_CMD_CAD_OFF
              (RB_DISABLE_CAD, 0).  CAD is disabled.  This means that the
              CAD keystroke will cause a SIGINT signal to be sent to init
              (process 1), whereupon this process may decide upon a proper
              action (maybe: kill all processes, sync, reboot).
       LINUX_REBOOT_CMD_CAD_ON
              (RB_ENABLE_CAD, 0x89abcdef).  CAD is enabled.  This means that
              the CAD keystroke will immediately cause the action associated
              with LINUX_REBOOT_CMD_RESTART.
       LINUX_REBOOT_CMD_HALT
              (RB_HALT_SYSTEM, 0xcdef0123; since Linux 1.1.76).  The message
              "System halted." is printed, and the system is halted.
              Control is given to the ROM monitor, if there is one.  If not
              preceded by a sync(2), data will be lost.
       LINUX_REBOOT_CMD_KEXEC
              (RB_KEXEC, 0x45584543, since Linux 2.6.13).  Execute a kernel
              that has been loaded earlier with kexec_load(2).  This option
              is available only if the kernel was configured with
              CONFIG_KEXEC.
       LINUX_REBOOT_CMD_POWER_OFF
              (RB_POWER_OFF, 0x4321fedc; since Linux 2.1.30).  The message
              "Power down." is printed, the system is stopped, and all power
              is removed from the system, if possible.  If not preceded by a
              sync(2), data will be lost.
       LINUX_REBOOT_CMD_RESTART
              (RB_AUTOBOOT, 0x1234567).  The message "Restarting system." is
              printed, and a default restart is performed immediately.  If
              not preceded by a sync(2), data will be lost.
       LINUX_REBOOT_CMD_RESTART2
              (0xa1b2c3d4; since Linux 2.1.30).  The message "Restarting
              system with command '%s'" is printed, and a restart (using the
              command string given in arg) is performed immediately.  If not
              preceded by a sync(2), data will be lost.
       LINUX_REBOOT_CMD_SW_SUSPEND
              (RB_SW_SUSPEND, 0xd000fce1; since Linux 2.5.18).  The system
              is suspended (hibernated) to disk.  This option is available
              only if the kernel was configured with CONFIG_HIBERNATION.
*/

type Page struct {
    SensorFile  []byte
    CalibrationFile  []byte
    ParametersFile []byte
    Message []byte
}

var templates = template.Must(template.ParseFiles("tmpl/dashboard.html",
                                                  "tmpl/settings.html",
                                                  "tmpl/success.html",
                                                  "tmpl/systemlogs.html",
                                                  "tmpl/reboot.html",
                                                  "tmpl/shutdown.html",))
var validPath = regexp.MustCompile("^/($|dashboard|settings|systemlogs|reboot|shutdown)?$")

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
  Param string
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "GET":
      p := &Page{}
      renderTemplate(w, "dashboard", p)
    case "POST":
      paramFile, err := loadFile("parameters.config")
      CheckErrSend(w, err, "Could not open parameters: %s")
      jsonData := dataAjax{Data: readMoister(), Param: fmt.Sprintf("%s", paramFile) }
      json.NewEncoder(w).Encode(jsonData)
    default:
      fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
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
        CheckErrSend(w, err, "Could not write config file %s")
        msg := fmt.Sprintf("File %s was saved.", fileName)
        renderTemplate(w, "success", &Page{Message: []byte(msg)})
    default:
        fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
    }

}

func systemLogsHandler(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "GET":
      logs, err := loadFile("logs")
      CheckErrSend(w, err, "Could not open logs: %s")
      p := &Page{Message: logs}
      renderTemplate(w, "systemlogs", p)
    case "POST":
        fmt.Fprintf(w, "Sorry, only GET methods are supported.")
    default:
        fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
  }
}

func rebootHandler(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "GET":
      p := &Page{}
      renderTemplate(w, "reboot", p)
    case "POST":
      fmt.Fprintf(w, "<p>Rebooting PoppyCon - Wait a few minutes and go back home <a href='/' > <- Dashboard</a></p>")
      time.Sleep(3 * time.Second)
      err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
      if err != nil{
        log.Println(err)
        fmt.Fprintf(w, "<p>Error, could not reboot <a href='/' > <- Dashboard</a></p>")
      }

    default:
        fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
  }
}

func shutdownHandler(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "GET":
      p := &Page{}
      renderTemplate(w, "shutdown", p)
    case "POST":
      fmt.Fprintf(w, "<p>Shutting down PoppyCon! You'll need to boot it up manually...</p>")
      time.Sleep(3 * time.Second)
      err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
      if err != nil{
        log.Println(err)
        fmt.Fprintf(w, "<p>Error, could not shutdown <a href='/' > <- Dashboard</a></p>")
      }

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


func main() {

    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    http.HandleFunc("/", makeHandler(dashboardHandler))
    http.HandleFunc("/settings", makeHandler(settingsHandler))
    http.HandleFunc("/systemlogs", makeHandler(systemLogsHandler))
    http.HandleFunc("/reboot", makeHandler(rebootHandler))
    http.HandleFunc("/shutdown", makeHandler(shutdownHandler))
    // http.HandleFunc("/settings/save", makeHandler(settingsSaveHandler))

    log.Fatal(http.ListenAndServe(":81", nil))
}
