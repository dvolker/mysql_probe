package statusserver

import (
  "fmt"
  "log"
  "net/http"
  "os"
  "regexp"
)

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func handler(w http.ResponseWriter, r *http.Request) {
  f, err := os.Open(r.URL.Path[1:])
  check(err)

  b1 := make([]byte, 10)
  n1, err := f.Read(b1)
  check(err)

  match, err := regexp.MatchString("up", string(b1))
  check(err)

  if match {
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "%s (read %d bytes)\n", string(b1), n1)
  } else {
    w.WriteHeader(http.StatusServiceUnavailable)
    fmt.Fprintf(w, "%s (read %d bytes)\n", string(b1), n1)
  }
}

func Run() {
  http.HandleFunc("/", handler)
  //fs := http.FileServer(http.Dir("tmp"))
  //http.Handle("/", fs)

  log.Println("Listening...")
  http.ListenAndServe(":3000", nil)
}