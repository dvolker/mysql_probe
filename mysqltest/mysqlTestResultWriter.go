package mysqltest

import (
  "os"
	"fmt"
	"path/filepath"
)

func TestResultPath(reportdir string, testname string) string {
  return fmt.Sprintf("%s.agent.txt", filepath.Join(reportdir, "/", testname))
}

func (t *MysqlTest) WriteTextResult(testname string, status string) {
	path := TestResultPath(t.reportdir, testname)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		//log.Fatal(err)
		desc := fmt.Sprintf("Couldn't open result file \"%s\": ", err.Error())
		t.JsonLog(desc)
		os.Stderr.WriteString(desc)
		return
	}
	defer file.Close()
	t.JsonLog(fmt.Sprintf("Test: %s result: %s", testname, status))
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
//		t.JsonLog(desc)
//		os.Stderr.WriteString(desc)
//		return
//	}
//	defer file.Close()
//	t.JsonLog(fmt.Sprintf("Test: %s result: %v: %s", testname, passed, description))
//	file.WriteString(response)
//}