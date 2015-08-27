package mysqltest

import (
  "os"
	"fmt"
	"path/filepath"
)

type MysqlTestResultWriter struct {
  test    *MysqlTest
  result  *MysqlTestResult
}

func TestResultPath(reportdir string, testname string) string {
  return fmt.Sprintf("%s.agent.txt", filepath.Join(reportdir, "/", testname))
}

// writes corresponding files for each test result entry
func (tw *MysqlTestResultWriter) WriteResult() {
  for k,v := range tw.result.Entries {
    tw.WriteTextResult(k,v)
  }
}

// write an individual test result into to a text file
func (tw *MysqlTestResultWriter) WriteTextResult(testname string, status string) {
	path := TestResultPath(tw.test.reportdir, testname)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		//log.Fatal(err)
		desc := fmt.Sprintf("Couldn't open result file \"%s\": ", err.Error())
		tw.test.JsonLog(desc)
		os.Stderr.WriteString(desc)
		return
	}
	defer file.Close()
	tw.test.JsonLog(fmt.Sprintf("Test: %s result: %s", testname, status))
	file.WriteString(status)
}

//func (t *MysqlTest) WriteHttpResult(testname string, passed bool, description string) {
//	status := "503 Service Unavailable"
//	if passed {
//		status = "200 OK"
//	}
//	now := time.Now().Format(time.RFC1123Z)
//	response := fmt.Sprintf("HTTP/1.1 %s\r\nDate: %s\r\nContent-Type: text/plain\r\n\r\n%s\r\n", status, now, description)
//
//	path := fmt.Sprintf("%s.http.txt", filepath.Join(t.reportdir, "/", testname))
//	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
//	if err != nil {
//		//log.Fatal(err)
//		desc := fmt.Sprintf("Couldn't open result file \"%s\": ", err.Error())
//		tw.test.JsonLog(desc)
//		os.Stderr.WriteString(desc)
//		return
//	}
//	defer file.Close()
//	tw.test.JsonLog(fmt.Sprintf("Test: %s result: %v: %s", testname, passed, description))
//	file.WriteString(response)
//}