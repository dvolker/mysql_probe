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

var required_up_checks = []string{"connect", "connection_count_lte_2400"}

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
    if is_up {
      is_up = s.testResultIsUp(testname)
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

func (s *StatuServer) testResultIsUp(testname string) bool {
  testpath := mysqltest.TestResultPath(s.reportdir, testname)
  match := false

  defer func() {
    if r := recover(); r != nil {
      log.Println("down via failure checking " + testpath + " (skipping all subsequent checks)")
    }
  }()

  f, err := os.Open(testpath)
  check(err)

  b1 := make([]byte, 10)
  _, err = f.Read(b1)
  check(err)

  match, err = regexp.MatchString("up", string(b1))
  check(err)

  if match {
    log.Println("up via " + testpath)
  } else {
    log.Println("down via " + testpath + " (skipping all subsequent checks)")
  }

  return match
}
