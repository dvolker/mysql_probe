package statusserver

import (
  "strconv"
  "fmt"
  "log"
  "net/http"
  "os"
  "regexp"
  "github.com/haikulearning/mysql_probe/mysqltest"
)

var required_up_checks = []string{"connection_count_lte_2400", "connect"}

type StatuServer struct {
	reportdir     string
	port          int
}

func StartStatuServer(reportdir string, port int) *StatuServer {
	s := StatuServer{reportdir: reportdir, port: port}

  s.Start()

	return &s
}

// Start up the status server
func (s *StatuServer) Start() {
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    s.handler(w, r)
  })
  //fs := http.FileServer(http.Dir("tmp"))
  //http.Handle("/", fs)

  log.Println("Listening to port " + strconv.Itoa(s.port))
  http.ListenAndServe(":" + strconv.Itoa(s.port), nil)
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func (s *StatuServer) handler(w http.ResponseWriter, r *http.Request) {
  is_up := true

  for _,testname := range required_up_checks {
    testpath := mysqltest.TestResultPath(s.reportdir, testname)
    log.Println("checking " + testpath)

    f, err := os.Open(testpath)
    if err != nil {
      is_up = false
    } else {
      b1 := make([]byte, 10)
      _, err := f.Read(b1)
      check(err)

      match, err := regexp.MatchString("up", string(b1))
      check(err)

      if !match && is_up {
        is_up = false
      }
    }
  }

  if is_up {
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "up\n")
  } else {
    w.WriteHeader(http.StatusServiceUnavailable)
    fmt.Fprintf(w, "down\n")
  }
}
