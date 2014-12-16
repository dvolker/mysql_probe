/*
   Copyright 2014 Derek Volker

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package mysqltest

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"time"
  "path/filepath"
)

type MysqlTest struct {
	Name     string
	filedirectory string
	host     string
	user     string
	port     int
	pass     string
	interval int
	timeout  int
	db       *sql.DB
}

func NewMysqlTest(name string, host string, port int, user string, pass string, interval int, timeout int, filedirectory string) *MysqlTest {
	m := MysqlTest{Name: name, host: host, port: port, user: user, pass: pass, interval: interval, timeout: timeout, filedirectory: filedirectory}
	return &m
}

func (t *MysqlTest) Run() {
	if t.interval <= 0 {
		panic("interval must be a positive integer")
	}
	// run checks on intervals
	for _ = range time.Tick(time.Duration(t.interval) * time.Millisecond) {
		// TODO: expire a test with a timeout
		t.RunOnce()
	}
}

// func writeHttpResult(filedirectory string) {
//
// }
func (t *MysqlTest) RunOnce() {
	err := t.Connect()
  if (err != nil) {
    t.WriteResult("connect", false, err.Error())
  }
	defer t.Disconnect()
}

func (t *MysqlTest) Disconnect() {
	fmt.Println("Disconnecting")
  if (t.db != nil) {
    t.db.Close()
  }
}

func (t *MysqlTest) WriteResult(testname string, passed bool, description string) {
  file, err := os.OpenFile(fmt.Sprintf("%s.txt", filepath.Join(t.filedirectory, "/", testname)),
      os.O_WRONLY | os.O_CREATE | os.O_TRUNC | os.O_EXCL, 0644)
  if err != nil {
    //log.Fatal(err)
    panic(err.Error())
  }
  defer file.Close()

  status := 500
  if (passed) {
    status = 200
  }

  file.WriteString(fmt.Sprintf("HTTP/1.1 %d\r\n\r\n%s", status, description))
}
func (t *MysqlTest) Connect() (error) {
	fmt.Println("Host: ", t.host, " Port: ", t.port, " User: ", t.user, " Pass: ", t.pass)

	// Create dsn like such https://github.com/Go-SQL-Driver/MySQL/#dsn-data-source-name
	// [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	// username:password@protocol(address)/dbname?param=value
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?timeout=%dms", t.user, t.pass, t.host, t.port, t.timeout)
	fmt.Println("DSN: ", dsn)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
    return err
	}

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
    return err
	}
	defer db.Close()

	fmt.Println("SUCCESS")
	t.db = db
  return nil
}
